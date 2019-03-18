package security

import (
	"fmt"
	"sort"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// User represents an authenicated user
type User struct {
	username string
	hash     string
	role     string
}

type key int

// Claims is the context key for user claims
const Claims key = 0

var (
	users     = []User{}
	secretKey []byte
)

// SetSecretKey sets the secret key
func SetSecretKey(sk string) {
	secretKey = []byte(sk)
}

// AddUser adds a new user
func AddUser(username, password, role string) (user User, error error) {
	if _, found := findUser(username); found {
		error = fmt.Errorf("security: username %s already exists", username)
	} else {
		if hash, error := generateKey(password); error == nil {
			user.username = username
			user.hash = hash
			user.role = role
			users = append(users, user)
			sort.Slice(users, func(i, j int) bool {
				return users[i].username < users[j].username
			})
		}
	}
	return
}

// VerifyUser authenticates a user
func VerifyUser(username, password string) (user User, error error) {
	var found bool
	if user, found = findUser(username); found {
		if compareKey(user.hash, password) != nil {
			error = fmt.Errorf("security: password does not match for username %s", username)
		}
	} else {
		error = fmt.Errorf("security: username %s does not exist", username)
	}
	return
}

// CreateToken creates a JSON Web Token for an authenticated user
func (user User) CreateToken() (tokenString string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.username,
		"role":     user.role,
	})
	tokenString, err = token.SignedString(secretKey)
	return
}

// VerifyToken verifies a JSON Web Token and return the claims
func VerifyToken(tokenValue string) (claims jwt.Claims, err error) {
	token, err := jwt.Parse(tokenValue, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("signing method is not valid")
		}
		return secretKey, nil
	})
	if err == nil {
		if token.Valid {
			claims = token.Claims
		} else {
			err = fmt.Errorf("invalid authorization token")
		}
	}
	return
}

func findUser(username string) (user User, found bool) {
	i := sort.Search(len(users), func(i int) bool {
		return users[i].username >= username
	})
	if i < len(users) && users[i].username == username {
		user = users[i]
		found = true
	}
	return
}

func generateKey(password string) (hash string, err error) {
	saltedBytes := []byte(password)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err == nil {
		hash = string(hashedBytes[:])
	}
	return
}

func compareKey(hash string, password string) error {
	incoming := []byte(password)
	existing := []byte(hash)
	return bcrypt.CompareHashAndPassword(existing, incoming)
}
