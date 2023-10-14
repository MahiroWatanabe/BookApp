package main

import (
	"context"
	"log"
	"next-go-app/ent"
	"next-go-app/ent/user"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {

	// PostgreSQLに接続
	client, err := ent.Open("postgres", "host=db port=5432 user=postgres dbname=db password=password sslmode=disable")

	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	router := gin.Default()

	// CORS設定
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	})

	// ルートハンドラの定義
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, Bookers!",
		})
	})

	// ユーザ新規登録機能
	router.POST("users/sign_up", func(c *gin.Context) {

		// リクエストボディにあるemailなどを構造体に入れる
		type SignUpRequest struct {
			Email    string `json:"email" binding:"required"`
			Name     string `json:"name" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		var req SignUpRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		newUser, err := client.User.
			Create().
			SetEmail(req.Email).
			SetName(req.Name).
			SetPassword(req.Password).
			Save(context.Background())

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error(), "messsage": "sign up missing"})
			return
		}
		c.JSON(201, gin.H{"user": newUser})

	})

	// ユーザログイン機能
	router.POST("users/sign_in", func(c *gin.Context) {

		type SignInRequest struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		var req SignInRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		sign_in_user, err := client.User.Query().
			Where(user.EmailEQ(req.Email), user.PasswordEQ(req.Password)).
			First(context.Background())

		if err != nil {
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}

		c.JSON(200, gin.H{"user": sign_in_user})

	})

	// 本の新規登録
	router.POST("/books", func(c *gin.Context) {

		// 本の新規登録で送られてくるリクエストを型定義
		type NewBookRequest struct {
			Title  string `json:"title" binding:"required"`
			Body   string `json:"body" binding:"required"`
			UserId int    `json:"user_id" binding:"required"`
		}

		// reqをNewBookRequestで定義
		var req NewBookRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		// 本の情報を保存
		newBook, err := client.Book.
			Create().
			SetTitle(req.Title).
			SetBody(req.Body).
			SetUserID(req.UserId).
			Save(context.Background())

		// エラーがある場合はエラーを返して終了
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error(), "message": "create book missing"})
			return
		}

		// 保存したBookの情報をレスポンスとして返す
		c.JSON(201, newBook)
	})

	// 本の一覧を取得
	router.GET("/books", func(c *gin.Context) {

		books, err := client.Book.Query().All(context.Background())

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error(), "message": "Could not get the book list."})
			return
		}

		c.JSON(200, books)
	})

	// 本の情報を取得
	router.GET("/books/:id", func(c *gin.Context) {

		bookIDStr := c.Param("id")

		// 文字->数字変換
		bookID, err := strconv.Atoi(bookIDStr)

		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid Book ID"})
			return
		}

		book, err := client.Book.Get(context.Background(), bookID)

		if err != nil {
			c.JSON(404, gin.H{"error": err.Error(), "message": "Book with specified id not found"})
			return
		}

		c.JSON(200, book)

	})

	// 本情報を更新する。
	router.PATCH("/books/:id", func(c *gin.Context) {

		type UpdateBookRequest struct {
			Title  string `json:"title" binding:"required"`
			Body   string `json:"body" binding:"required"`
			UserId int    `json:"user_id" binding:"required"`
		}

		var book UpdateBookRequest

		if err := c.ShouldBindJSON(&book); err != nil {
			c.JSON(400, gin.H{"error": err.Error(), "message": "Invalid Book ID"})
			return
		}

		bookIDStr := c.Param("id")

		bookID, err := strconv.Atoi(bookIDStr)

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error(), "message": "could not translation string->int"})
			return
		}
		update_book, err := client.Book.
			UpdateOneID(bookID).
			SetTitle(book.Title).
			SetBody(book.Body).
			Save(context.Background())

		if err != nil {
			c.JSON(404, gin.H{"error": err.Error(), "message": "Couldn't update"})
			return
		}

		c.JSON(200, update_book)
	})

	// 本を削除
	router.DELETE("/books/:id", func(c *gin.Context) {
		bookIDStr := c.Param("id")

		bookID, err := strconv.Atoi(bookIDStr)

		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid Book ID"})
			return
		}

		err = client.Book.DeleteOneID(bookID).Exec(context.Background())

		if err != nil {
			c.JSON(404, gin.H{"error": "Failed to delete"})
			return
		}

		c.JSON(200, gin.H{"message": "Delete completed"})
	})

	// サーバーの開始
	router.Run(":8080")
}
