package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/bener07/gled-maintainer/"
)


type Data struct {
    ID  int    `json:"id"`
    Name string `json:"name"`
}


// homeHandler é o manipulador para a rota "/"
func homeHandler(c *gin.Context) {
    db, err := db.ConnectDB()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Erro ao conectar à base de dados",
        })
        return
    }
    defer db.Close()

    data, err := db.json(db, "SELECT * FROM data", Data);
    c.JSON(http.StatusOK, data)
}



func main(){
    r := gin.Default()

    r.GET("/", homeHandler)

    // App Listener
    r.Run(":8000")

    http.Handle("/", r)
}
