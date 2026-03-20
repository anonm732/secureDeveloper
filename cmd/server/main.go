package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	_ "modernc.org/sqlite"
)

const authorizationCookieName = "authorization"

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"-"`
	Salt     int    `json:"-"`
	Balance  int64  `json:"balance"`
	IsAdmin  bool   `json:"is_admin"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type WithdrawAccountRequest struct {
	Password string `json:"password"`
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Balance  int64  `json:"balance"`
	IsAdmin  bool   `json:"is_admin"`
}

type LoginResponse struct {
	AuthMode string       `json:"auth_mode"`
	Token    string       `json:"token"`
	User     UserResponse `json:"user"`
}

type PostView struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	OwnerID     uint   `json:"owner_id"`
	Author      string `json:"author"`
	AuthorEmail string `json:"author_email"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type PostListResponse struct {
	Posts []PostView `json:"posts"`
}

type PostResponse struct {
	Post PostView `json:"post"`
}

type DepositRequest struct {
	Amount int64 `json:"amount"`
}

type BalanceWithdrawRequest struct {
	Amount int64 `json:"amount"`
}

type TransferRequest struct {
	ToUsername string `json:"to_username"`
	Amount     int64  `json:"amount"`
}

type Store struct {
	db *sql.DB
}

type SessionStore struct {
	tokens map[string]User
}

const app_db string = "./ext/db/sqlite/app.db"
const scheme_sql string = "./ext/db/sqlite/init/schema.sql"
const seed_sql string = "./ext/db/sqlite/init/seed.sql"

const log_dir string = "./logs/"

func main() {
	store, err := openStore(app_db, scheme_sql, seed_sql)
	if err != nil {
		panic(err)
	}
	defer store.close()

	initLogger()

	sessions := newSessionStore()

	router := gin.Default()
	router.Use(JSONLogger())
	registerStaticRoutes(router)

	auth := router.Group("/api/auth")
	{
		auth.POST("/register", func(c *gin.Context) {
			var request RegisterRequest
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "invalid register request"})
				return
			}

			user, ok, err := store.findUserByUsername(request.Username)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
				return
			}

			if ok || user.Username == request.Username {
				c.JSON(http.StatusConflict, gin.H{"message": "username is already used."})
				return
			}
			// email 형식 검사

			// insertUser 호출
			name, ok, err := store.insertUser(request)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to insert user"})
			}

			c.JSON(http.StatusAccepted, gin.H{
				"message": "Register Success",
				"user": gin.H{
					"username": request.Username,
					"name":     name,
					"email":    request.Email,
					"phone":    request.Phone,
				},
			})
		})

		auth.POST("/login", func(c *gin.Context) {
			var request LoginRequest
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "invalid login request"})
				return
			}

			user, ok, err := store.findUserByUsername(request.Username)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load user"})
				return
			}

			password, _ := saltingfPassword(request.Password, user.Salt)
			if !ok || user.Password != password {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
				return
			}

			token, err := sessions.create(user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create session"})
				return
			}

			c.SetSameSite(http.SameSiteLaxMode)
			c.SetCookie(authorizationCookieName, token, 60*60*8, "/", "", false, true)
			c.JSON(http.StatusOK, LoginResponse{
				AuthMode: "header-and-cookie",
				Token:    token,
				User:     makeUserResponse(user),
			})
		})

		auth.POST("/logout", func(c *gin.Context) {
			token := tokenFromRequest(c)
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
				return
			}
			if _, ok := sessions.lookup(token); !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization token"})
				return
			}

			sessions.delete(token)
			clearAuthorizationCookie(c)
			c.JSON(http.StatusOK, gin.H{
				"message": "logout success",
				"todo":    "replace with revoke or audit logic if needed",
			})
		})

		auth.POST("/withdraw", func(c *gin.Context) {
			var request WithdrawAccountRequest
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "invalid withdraw request"})
				return
			}

			token := tokenFromRequest(c)
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
				return
			}
			user, ok := sessions.lookup(token)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization token"})
				return
			}

			c.JSON(http.StatusAccepted, gin.H{
				"message": "dummy withdraw handler",
				"todo":    "replace with password check and account delete logic",
				"user":    makeUserResponse(user),
			})
		})
	}

	protected := router.Group("/api")
	{
		protected.GET("/me", func(c *gin.Context) {
			token := tokenFromRequest(c)
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
				return
			}
			user, ok := sessions.lookup(token)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization token"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"user": makeUserResponse(user)})
		})

		protected.POST("/banking/deposit", func(c *gin.Context) {
			var request DepositRequest
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "invalid deposit request"})
				return
			}

			token := tokenFromRequest(c)
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
				return
			}
			user, ok := sessions.lookup(token)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization token"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "dummy deposit handler",
				"todo":    "replace with balance increment query",
				"user":    makeUserResponse(user),
				"amount":  request.Amount,
			})
		})

		protected.POST("/banking/withdraw", func(c *gin.Context) {
			var request BalanceWithdrawRequest
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "invalid withdraw request"})
				return
			}

			token := tokenFromRequest(c)
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
				return
			}
			user, ok := sessions.lookup(token)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization token"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "dummy withdraw handler",
				"todo":    "replace with balance check and decrement query",
				"user":    makeUserResponse(user),
				"amount":  request.Amount,
			})
		})

		protected.POST("/banking/transfer", func(c *gin.Context) {
			var request TransferRequest
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "invalid transfer request"})
				return
			}

			token := tokenFromRequest(c)
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
				return
			}
			user, ok := sessions.lookup(token)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization token"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "dummy transfer handler",
				"todo":    "replace with transfer transaction and balance checks",
				"user":    makeUserResponse(user),
				"target":  request.ToUsername,
				"amount":  request.Amount,
			})
		})

		protected.GET("/posts", func(c *gin.Context) {
			token := tokenFromRequest(c)
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
				return
			}
			if _, ok := sessions.lookup(token); !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization token"})
				return
			}

			// posts table SELECT 로직 필요

			// 표시할 개수 제한 필요
			c.JSON(http.StatusOK, PostListResponse{
				Posts: []PostView{
					{
						ID:          1,
						Title:       "Dummy Post",
						Content:     "This is a fixed dummy response. Replace this later with real board logic.",
						OwnerID:     1,
						Author:      "Alice Admin",
						AuthorEmail: "alice.admin@example.com",
						CreatedAt:   "2026-03-19T09:00:00Z",
						UpdatedAt:   "2026-03-19T09:00:00Z",
					},
				},
			})
		})

		protected.POST("/posts", func(c *gin.Context) {
			var request CreatePostRequest
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "invalid create request"})
				return
			}

			token := tokenFromRequest(c)
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
				return
			}
			user, ok := sessions.lookup(token)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization token"})
				return
			}

			// posts table INSERT 로직 필요
			ok, err = store.createPost(request, user)

			now := time.Now().Format(time.RFC3339)
			c.JSON(http.StatusCreated, gin.H{
				"message": "post uploaded",
				"post": PostView{
					ID:          1,
					Title:       strings.TrimSpace(request.Title),
					Content:     strings.TrimSpace(request.Content),
					OwnerID:     user.ID,
					Author:      user.Name,
					AuthorEmail: user.Email,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			})
		})

		protected.GET("/posts/:id", func(c *gin.Context) {
			token := tokenFromRequest(c)
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
				return
			}
			if _, ok := sessions.lookup(token); !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization token"})
				return
			}

			id, err := strconv.Atoi(c.Param("id"))
			if id < 1 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
				return
			}

			// posts table SELECT ~ WHERE id = ? 로직 필요
			post, _, err := store.readPost(id)
			if err != nil {
				c.JSON(http.StatusNoContent, gin.H{"message": "content not exsited"})
				return
			}
			fmt.Println("post.Title:", post.Title)

			c.JSON(http.StatusOK, PostResponse{
				Post: PostView{
					post.ID,
					post.Title,
					post.Content,
					post.OwnerID,
					post.Author,
					post.AuthorEmail,
					post.CreatedAt,
					post.UpdatedAt,
				},
			})
		})

		protected.PUT("/posts/:id", func(c *gin.Context) {
			var request UpdatePostRequest
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "invalid update request"})
				return
			}

			token := tokenFromRequest(c)
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
				return
			}
			user, ok := sessions.lookup(token)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization token"})
				return
			}

			// posts table UPDATE TABLE ~ SET ~ WHERE id = ? 로직 필요

			now := time.Now().Format(time.RFC3339)
			c.JSON(http.StatusOK, gin.H{
				"message": "dummy update post handler",
				"todo":    "replace with ownership check and update query",
				"post": PostView{
					ID:          1,
					Title:       strings.TrimSpace(request.Title),
					Content:     strings.TrimSpace(request.Content),
					OwnerID:     user.ID,
					Author:      user.Name,
					AuthorEmail: user.Email,
					CreatedAt:   "2026-03-19T09:00:00Z",
					UpdatedAt:   now,
				},
			})
		})

		protected.DELETE("/posts/:id", func(c *gin.Context) {
			token := tokenFromRequest(c)
			if token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
				return
			}
			if _, ok := sessions.lookup(token); !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization token"})
				return
			}

			// posts table UPDATE TABLE posts SET <삭제 여부> WHERE id = ? 로직 필요

			c.JSON(http.StatusOK, gin.H{
				"message": "dummy delete post handler",
				"todo":    "replace with ownership check and delete query",
			})
		})
	}

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}

// --- main() func end

func openStore(databasePath, schemaFile, seedFile string) (*Store, error) {
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	store := &Store{db: db}
	if err := store.initialize(schemaFile, seedFile); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) close() error {
	return s.db.Close()
}

func (s *Store) initialize(schemaFile, seedFile string) error {
	if err := s.execSQLFile(schemaFile); err != nil {
		return err
	}
	if err := s.execSQLFile(seedFile); err != nil {
		return err
	}
	return nil
}

func (s *Store) execSQLFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(string(content))
	return err
}

// --- no need to modify functions above

func (s *Store) findUserByUsername(username string) (User, bool, error) {
	row := s.db.QueryRow(`
		SELECT id, username, name, email, phone, password, salt, balance, is_admin
		FROM users
		WHERE username = ?
	`, strings.TrimSpace(username))

	var user User
	var isAdmin int64
	if err := row.Scan(&user.ID, &user.Username, &user.Name, &user.Email, &user.Phone, &user.Password, &user.Salt, &user.Balance, &isAdmin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, false, nil
		}
		return User{}, false, err
	}
	user.IsAdmin = isAdmin == 1

	return user, true, nil
}

// added
func saltingfPassword(password string, salt int) (string, int) {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	s := 0
	if salt == 0 {
		s = rand.Intn(999999)
	} else {
		s = salt
	}
	strSalt := strconv.Itoa(s)
	saltedPw := password + strSalt

	hash := sha1.New()
	hash.Write([]byte(saltedPw))
	byteHash := hash.Sum(nil)
	strHash := fmt.Sprintf("%x", byteHash)

	return strHash, s
}

// added
func (s *Store) insertUser(user RegisterRequest) (string, bool, error) {
	saltedPw, salt := saltingfPassword(user.Username, 0)

	insertQuery := `
		INSERT INTO users (username, name, email, phone, password, salt, balance, is_admin)
		VALUES (?, ?, ?, ?, ?, ?, 0, 0);
	`

	if _, err := s.db.Exec(insertQuery,
		strings.TrimSpace(user.Username),
		strings.TrimSpace(user.Name),
		strings.TrimSpace(user.Email),
		strings.TrimSpace(user.Phone),
		saltedPw,
		salt); err != nil {
		fmt.Println("db insert failed:", err)
		return "", false, err
	}

	return user.Name, true, nil
}

func (s *Store) createPost(reqPost CreatePostRequest, user User) (bool, error) {
	insertQuery := `
		INSERT INTO posts (title, content, owner_id, author, author_email)
		VALUES (?, ?, ?, ?, ?)
	`
	if _, err := s.db.Exec(insertQuery,
		reqPost.Title,
		reqPost.Content,
		user.ID,
		user.Name,
		user.Email,
	); err != nil {
		fmt.Println("err:", err)
		return false, err
	}

	return true, nil
}

func (s *Store) readPost(postId int) (PostView, bool, error) {
	row := s.db.QueryRow(`
		SELECT id, title, content, owner_id, author, author_email, created_at, updated_at
		FROM posts
		WHERE id = ?
	`, postId)

	var post PostView
	if err := row.Scan(&post.ID, &post.Title, &post.Content, &post.OwnerID, &post.Author, &post.AuthorEmail, &post.CreatedAt, &post.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("errorsIs", sql.ErrNoRows)
			return PostView{}, false, nil
		}
		fmt.Println("err : ", err)
		return PostView{}, false, err
	}
	fmt.Println("readPost")

	return post, true, nil
}

func initLogger() {
	if _, err := os.ReadDir(log_dir); err != nil {
		os.MkdirAll(log_dir, 0755)
		fmt.Println("New logs directory is created.")
	}
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(&lumberjack.Logger{
		Filename:   log_dir + "api_access.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	})
}

func JSONLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.WithFields(log.Fields{
			"ip":     c.ClientIP(),
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"query":  c.Request.URL.RawQuery,
			"header": c.Request.Header,
		})
		c.Next()
	}
}

// --- no need to modify functions below

func newSessionStore() *SessionStore {
	return &SessionStore{
		tokens: make(map[string]User),
	}
}

func (s *SessionStore) create(user User) (string, error) {
	token, err := newSessionToken()
	if err != nil {
		return "", err
	}

	s.tokens[token] = user
	return token, nil
}

func (s *SessionStore) lookup(token string) (User, bool) {
	user, ok := s.tokens[token]
	return user, ok
}

func (s *SessionStore) delete(token string) {
	delete(s.tokens, token)
}

// fe 페이지 캐싱으로 테스트에 혼동이 있어, 별도 처리없이 main에 두시면 될 것 같습니다
// registerStaticRoutes 는 정적 파일(HTML, JS, CSS)을 제공하는 라우트를 등록한다.
func registerStaticRoutes(router *gin.Engine) {
	// 브라우저 캐시 비활성화 — 정적 파일과 루트 경로에만 적용
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/static/") || c.Request.URL.Path == "/" {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	})
	router.Static("/static", "./static")
	router.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})
}

func makeUserResponse(user User) UserResponse {
	return UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Name:     user.Name,
		Email:    user.Email,
		Phone:    user.Phone,
		Balance:  user.Balance,
		IsAdmin:  user.IsAdmin,
	}
}

func clearAuthorizationCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(authorizationCookieName, "", -1, "/", "", false, true)
}

func tokenFromRequest(c *gin.Context) string {
	headerValue := strings.TrimSpace(c.GetHeader("Authorization"))
	if headerValue != "" {
		return headerValue
	}

	cookieValue, err := c.Cookie(authorizationCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookieValue)
}

func newSessionToken() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}
