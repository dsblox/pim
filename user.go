package main

import (
	"fmt"
	"net/mail"
	"github.com/satori/go.uuid"
	"github.com/dgrijalva/jwt-go"
	"time"
)

// Create the JWT key used to create the signature
var jwtKey = []byte("pim_secret_key_to_be_more_secure")

// Create a struct that will be encoded to a JWT.
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type User struct {
   id string        // unique id for this user
   name string      // name of the user
   email string     // email address of the user
   password string  // unencrypted (for now) password
}

// Create a struct to read the username and password from a request body
type UserCredentials struct {
    Password string `json:"password"`
    Email    string `json:"email"`
}

func UserGetAuthToken(username string, expirationTime time.Time) (string, error) {

   // create the JWT claims, which includes username and expiry time
   claims := &Claims{
   	Username: username,
      StandardClaims: jwt.StandardClaims{
	      // In JWT, the expiry time is expressed as unix milliseconds
   	   ExpiresAt: expirationTime.Unix(),
      },
   }

	// declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create the JWT string or any error generated
	return token.SignedString(jwtKey)
}

func UserValidateAuthToken(tknStr string) (string, PimErrId) {

	// Initialize a new instance of `Claims`
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
	  return jwtKey, nil
	})
	if err != nil {
	   if err == jwt.ErrSignatureInvalid {
	       // w.WriteHeader(http.StatusUnauthorized)
	       return "", authSig
	   }
	   // w.WriteHeader(http.StatusBadRequest)
	   return "", authFail
	}
	if !tkn.Valid {
	   // w.WriteHeader(http.StatusUnauthorized)
	   return "", authToken
	}

	// if we got here the token is valid so return the user
	return claims.Username, success
}


/*
===============================================================================
 User
-------------------------------------------------------------------------------
 Basic creator function and getter / setters for a user.
-----------------------------------------------------------------------------*/
func NewUser(newName string, newEmail string, newPassword string) (*User, PimErrId) {
	id := uuid.NewV4()
	// TBD - make sure password meets minimum standards
	// ...will need a new error for this
	noob := &User{id:id.String(), name:newName, email:"", password:newPassword}
	err := noob.SetEmail(newEmail)
	if err != nil {
		return nil, authBadEmail
	} 
	return noob, success
}


func (u User) String() string {
	return fmt.Sprintf("%s (%s)",u.name, u.email)
}
func (u *User) SetId(newId string) {
	u.id = newId
}
func (u *User) GetId() string {
	return u.id
}
func (u *User) SetName(newName string) {
	u.name = newName
}
func (u *User) GetName() string {
	return u.name
}
func (u *User) SetEmail(newEmail string) error {
	_, err := mail.ParseAddress(newEmail)
	if (err == nil) {
		u.email = newEmail
		fmt.Printf("SetEmail(): good email = %s\n", newEmail)		
	} else {
		fmt.Printf("SetEmail(): bad email = %s\n", newEmail)
	}
	return err
}
func (u *User) GetEmail() string {
	return u.email
}
func (u *User) SetPassword(newPassword string) {
	u.password = newPassword
}
func (u *User) GetPassword() string {
	return u.password
}


/*
===============================================================================
 Users
-------------------------------------------------------------------------------
 Simple class to group users and help clients easily find them by email
 address or username.
-----------------------------------------------------------------------------*/
type Users []*User

func (list Users) FindByEmail(email string) *User {
	fmt.Printf("Users.FindByEmail(): listlen = %v\n", len(list))
	for _, curr := range list {
		fmt.Printf("Users.FindByEmail(): curr = %s\n", curr)
		if email == curr.GetEmail() {
			return curr
		}
	}
	return nil
}


