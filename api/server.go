package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "api/database"
)

type UserData struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}   

func Response(c *gin.Context, status int, data []byte) {
    c.Data(status, "application/json", data)
}

// homeHandler é o manipulador para a rota "/"
func homeHandler(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "API em execução",
    })
}

func returnData(c *gin.Context, query string) {
    db, err := database.ConnectDB()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Erro ao conectar à base de dados",
        })
        return
    }
    defer db.Close()

    data, err := database.GetQuery(db, query);
    Response(c, http.StatusOK, data)

    return
}

func main(){
    r := gin.Default()

    r.GET("/", homeHandler)

    // r.GET("/users/:id", func(c *gin.Context) {
    //     id := c.Param("id")
    //     returnData(c, "SELECT * FROM users WHERE id = ?")
    // })

    users := r.Group("/users")

    users.GET("/", func(c *gin.Context) {
        returnData(c, "SELECT * FROM users")
    })

    users.POST("/", func(c *gin.Context) {
        var data UserData

        if err := c.BindJSON(&data); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "id":   data.ID,
            "name": data.Name,
        })
    })

    // App Listener
    r.Run(":8000")

    http.Handle("/", r)
}
