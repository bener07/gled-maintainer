package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"api/database"

	"github.com/gin-gonic/gin"
)

// ─── Helpers ────────────────────────────────────────────────────────────────

func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func respond(c *gin.Context, code int, success bool, message string, data interface{}) {
	c.JSON(code, gin.H{"success": success, "message": message, "data": data})
}

// ─── Middleware ──────────────────────────────────────────────────────────────

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-API-Key")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func apiKeyMiddleware() gin.HandlerFunc {
	secret := os.Getenv("API_SECRET")
	return func(c *gin.Context) {
		// If no secret configured, skip auth (development mode)
		if secret == "" {
			c.Next()
			return
		}
		key := c.GetHeader("X-API-Key")
		if key != secret {
			respond(c, http.StatusUnauthorized, false, "Chave de API inválida", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

// ─── Route Setup ─────────────────────────────────────────────────────────────

func setupRoutes(r *gin.Engine, db *sql.DB) {
	// Public
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().Format(time.RFC3339)})
	})

	protected := r.Group("/")
	protected.Use(apiKeyMiddleware())

	// ── Stats ────────────────────────────────────────────────────────────────
	protected.GET("/stats", func(c *gin.Context) {
		totalClients, _  := database.QueryRow(db, "SELECT COUNT(*) as total FROM clients")
		onlineClients, _ := database.QueryRow(db, "SELECT COUNT(*) as total FROM clients WHERE status = 'online'")
		totalUpdates, _  := database.QueryRow(db, "SELECT COUNT(*) as total FROM updates")
		pendingUpdates, _:= database.QueryRow(db, "SELECT COUNT(*) as total FROM updates WHERE status = 'pending'")

		respond(c, http.StatusOK, true, "", gin.H{
			"total_clients":   totalClients["total"],
			"online_clients":  onlineClients["total"],
			"total_updates":   totalUpdates["total"],
			"pending_updates": pendingUpdates["total"],
		})
	})

	// ── Clients ──────────────────────────────────────────────────────────────
	clients := protected.Group("/clients")

	clients.GET("/", func(c *gin.Context) {
		data, err := database.QueryRows(db,
			"SELECT id, name, hostname, os, version, status, last_heartbeat, created_at FROM clients ORDER BY id DESC")
		if err != nil {
			respond(c, http.StatusInternalServerError, false, "Erro ao obter clientes", nil)
			return
		}
		var result []interface{}
		json.Unmarshal(data, &result)
		respond(c, http.StatusOK, true, "", result)
	})

	clients.POST("/register", func(c *gin.Context) {
		var body struct {
			Name     string `json:"name"     binding:"required"`
			Hostname string `json:"hostname" binding:"required"`
			OS       string `json:"os"`
			Version  string `json:"version"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			respond(c, http.StatusBadRequest, false, "Dados inválidos: "+err.Error(), nil)
			return
		}

		// Return existing api_key if hostname already registered
		existing, _ := database.QueryRow(db, "SELECT id, api_key FROM clients WHERE hostname = ?", body.Hostname)
		if existing != nil {
			respond(c, http.StatusOK, true, "Cliente já registado", gin.H{
				"api_key": existing["api_key"], "client_id": existing["id"],
			})
			return
		}

		apiKey, err := generateAPIKey()
		if err != nil {
			respond(c, http.StatusInternalServerError, false, "Erro ao gerar chave de API", nil)
			return
		}

		osVal := body.OS
		if osVal == "" { osVal = "unknown" }
		ver := body.Version
		if ver == "" { ver = "0.0.0" }

		id, err := database.Exec(db,
			"INSERT INTO clients (name, hostname, os, version, api_key, status) VALUES (?, ?, ?, ?, ?, 'online')",
			body.Name, body.Hostname, osVal, ver, apiKey,
		)
		if err != nil {
			respond(c, http.StatusInternalServerError, false, "Erro ao registar cliente", nil)
			return
		}
		respond(c, http.StatusCreated, true, "Cliente registado com sucesso", gin.H{
			"client_id": id, "api_key": apiKey,
		})
	})

	clients.GET("/:id", func(c *gin.Context) {
		row, err := database.QueryRow(db,
			"SELECT id, name, hostname, os, version, status, last_heartbeat, created_at FROM clients WHERE id = ?",
			c.Param("id"))
		if err != nil || row == nil {
			respond(c, http.StatusNotFound, false, "Cliente não encontrado", nil)
			return
		}
		respond(c, http.StatusOK, true, "", row)
	})

	clients.DELETE("/:id", func(c *gin.Context) {
		n, err := database.ExecAffected(db, "DELETE FROM clients WHERE id = ?", c.Param("id"))
		if err != nil || n == 0 {
			respond(c, http.StatusNotFound, false, "Cliente não encontrado", nil)
			return
		}
		respond(c, http.StatusOK, true, "Cliente removido com sucesso", nil)
	})

	clients.POST("/:id/heartbeat", func(c *gin.Context) {
		var body struct {
			Version string `json:"version"`
			OS      string `json:"os"`
		}
		c.ShouldBindJSON(&body)

		var query string
		var args []interface{}
		now := time.Now()

		if body.Version != "" {
			query = "UPDATE clients SET last_heartbeat = ?, status = 'online', version = ?, updated_at = NOW() WHERE id = ?"
			args = []interface{}{now, body.Version, c.Param("id")}
		} else {
			query = "UPDATE clients SET last_heartbeat = ?, status = 'online', updated_at = NOW() WHERE id = ?"
			args = []interface{}{now, c.Param("id")}
		}

		n, err := database.ExecAffected(db, query, args...)
		if err != nil || n == 0 {
			respond(c, http.StatusNotFound, false, "Cliente não encontrado", nil)
			return
		}
		respond(c, http.StatusOK, true, "Heartbeat recebido", nil)
	})

	// ── Updates ──────────────────────────────────────────────────────────────
	updates := protected.Group("/updates")

	updates.GET("/", func(c *gin.Context) {
		data, err := database.QueryRows(db,
			"SELECT id, version, changelog, status, scheduled_at, created_at FROM updates ORDER BY id DESC")
		if err != nil {
			respond(c, http.StatusInternalServerError, false, "Erro ao obter atualizações", nil)
			return
		}
		var result []interface{}
		json.Unmarshal(data, &result)
		respond(c, http.StatusOK, true, "", result)
	})

	updates.POST("/", func(c *gin.Context) {
		var body struct {
			Version     string `json:"version"      binding:"required"`
			Changelog   string `json:"changelog"    binding:"required"`
			ScheduledAt string `json:"scheduled_at"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			respond(c, http.StatusBadRequest, false, "Dados inválidos: "+err.Error(), nil)
			return
		}

		var scheduledAt interface{}
		if body.ScheduledAt != "" {
			if t, err := time.Parse(time.RFC3339, body.ScheduledAt); err == nil {
				scheduledAt = t
			}
		}

		id, err := database.Exec(db,
			"INSERT INTO updates (version, changelog, scheduled_at) VALUES (?, ?, ?)",
			body.Version, body.Changelog, scheduledAt,
		)
		if err != nil {
			respond(c, http.StatusInternalServerError, false, "Erro ao criar atualização", nil)
			return
		}
		respond(c, http.StatusCreated, true, "Atualização criada com sucesso", gin.H{"id": id})
	})

	// Latest pending update — used by clients when polling
	updates.GET("/latest", func(c *gin.Context) {
		row, err := database.QueryRow(db,
			"SELECT id, version, changelog, scheduled_at FROM updates WHERE status = 'pending' ORDER BY id DESC LIMIT 1")
		if err != nil || row == nil {
			respond(c, http.StatusOK, true, "Sem atualizações pendentes", nil)
			return
		}
		respond(c, http.StatusOK, true, "", row)
	})

	updates.GET("/:id", func(c *gin.Context) {
		row, err := database.QueryRow(db,
			"SELECT id, version, changelog, status, scheduled_at, created_at FROM updates WHERE id = ?",
			c.Param("id"))
		if err != nil || row == nil {
			respond(c, http.StatusNotFound, false, "Atualização não encontrada", nil)
			return
		}
		respond(c, http.StatusOK, true, "", row)
	})

	updates.DELETE("/:id", func(c *gin.Context) {
		n, err := database.ExecAffected(db, "DELETE FROM updates WHERE id = ?", c.Param("id"))
		if err != nil || n == 0 {
			respond(c, http.StatusNotFound, false, "Atualização não encontrada", nil)
			return
		}
		respond(c, http.StatusOK, true, "Atualização removida", nil)
	})

	updates.POST("/:id/confirm", func(c *gin.Context) {
		updateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			respond(c, http.StatusBadRequest, false, "ID inválido", nil)
			return
		}
		var body struct {
			ClientID int    `json:"client_id" binding:"required"`
			Status   string `json:"status"    binding:"required"`
			Message  string `json:"message"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			respond(c, http.StatusBadRequest, false, "Dados inválidos: "+err.Error(), nil)
			return
		}
		if body.Status != "success" && body.Status != "failed" {
			respond(c, http.StatusBadRequest, false, "Status deve ser 'success' ou 'failed'", nil)
			return
		}

		_, err = database.Exec(db,
			`INSERT INTO update_confirmations (update_id, client_id, status, message) VALUES (?, ?, ?, ?)
			 ON DUPLICATE KEY UPDATE status=VALUES(status), message=VALUES(message)`,
			updateID, body.ClientID, body.Status, body.Message,
		)
		if err != nil {
			respond(c, http.StatusInternalServerError, false, "Erro ao registar confirmação", nil)
			return
		}

		// Mark update as applied when all clients report success
		database.Exec(db,
			`UPDATE updates SET status = 'applied' WHERE id = ?
			 AND (SELECT COUNT(*) FROM update_confirmations WHERE update_id = ? AND status = 'success')
			     >= (SELECT COUNT(*) FROM clients)`,
			updateID, updateID,
		)

		respond(c, http.StatusOK, true, "Confirmação registada", nil)
	})
}

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Erro ao conectar à base de dados: %v", err)
	}
	defer db.Close()

	if err := database.InitSchema(db); err != nil {
		log.Fatalf("Erro ao inicializar schema: %v", err)
	}

	r := gin.Default()
	r.Use(corsMiddleware())
	setupRoutes(r, db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("Maintainer API a correr na porta :%s", port)
	r.Run(":" + port)
}
