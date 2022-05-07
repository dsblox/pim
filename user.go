package main

import (
   "fmt"
   "log"
   "context"
   "net/mail"
   "net/http"
   "github.com/satori/go.uuid"
   "github.com/dgrijalva/jwt-go"
   "time"
   "golang.org/x/crypto/bcrypt"
   "errors"
)



/*
===============================================================================
 UserAuthentication - JWT Layer
-------------------------------------------------------------------------------
 This section of the file wraps and customizes the JWT package to create and
 authenticate access tokens.
-----------------------------------------------------------------------------*/

// Create the JWT key used to create the signature
var jwtKey = []byte("pim_secret_key_to_be_more_secure")

// set the expiration timeout on any created tokens
const jwtTimeout = (5 * time.Minute)

// Create a struct that will be encoded to a JWT.
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type Claims struct {
   Username string `json:"username"`
   jwt.StandardClaims
}

func UserGetAuthToken(username string, expirationTime time.Time) (string, error) {

   // create the JWT claims, which includes username and expiry time
   claims := &Claims{
      Username: username,
      StandardClaims: jwt.StandardClaims{
         // In JWT, the expiry time is expressed as unix milliseconds
         // note we take this as an argument so the time can be created
         // 1x in the HTTP layer - but for security we may want to move
         // it in here and force HTTP to keep it the same.
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

   // if we got here the token is valid so return the username
   return claims.Username, success
}

/*
===============================================================================
 UserAuthentication - User Layer
-------------------------------------------------------------------------------
 This section of the file represents the user objects themselves, which is an
 unique id, email and password for each user, and a few list functions on
 user slices.  Note the persistance interface is built over TaskDataMapper
 which we should rename PIMDataMapper.
-----------------------------------------------------------------------------*/
type User struct {
   id string        // unique id for this user
   name string      // name of the user
   email string     // email address of the user
   password []byte  // encrypted password
   persist TaskDataMapper // interface to store the user

}

// Create a struct to read the username and password from a request body
type UserCredentials struct {
    Password string `json:"password"`
    Email    string `json:"email"`
}

/*
===============================================================================
 User
-------------------------------------------------------------------------------
 Basic creator function and getter / setters for a user.
-----------------------------------------------------------------------------*/
func NewUser(newId string, newName string, newEmail string, newPassword string, storage TaskDataMapper) (*User, PimErrId) {

   // if id is provided we're loading from DB, if not, we're creating a user in memory   
   if len(newId) == 0 {
     newId = uuid.NewV4().String()
   }

   // create the user - we set the email after to check for valid email
   noob := &User{id:newId, name:newName, email:"", password:nil, persist:storage}
   err := noob.SetEmail(newEmail)
   if err != nil {
      return nil, authBadEmail
   }
   err = noob.SetNewPassword(newPassword)
   if err != nil {
      return nil, authBadEmail // TBD - new error code for bad password
   }
   return noob, success
}

func LoadUser(loadId string, loadName string, loadEmail string, loadPassword string, storage TaskDataMapper) (*User, PimErrId) {

   // this is intended to load from storage so must have an ID
   if len(loadId) == 0 {
    return nil, badRequest
   }

   // may want to check other things like that the pasword we are loading
   // is actually an encrypted string (how?) - or do we trust all callers
   // that this data was already validated?

   // create the user - we set the email after to check for valid email
   noob := &User{id:loadId, name:loadName, email:loadEmail, password:[]byte(loadPassword), persist:storage}
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
   }
   return err
}
func (u *User) GetEmail() string {
   return u.email
}
func (u *User) SetNewPassword(newPassword string) error {

  // Salt and hash the password using the bcrypt algorithm
  // The second argument is the cost of hashing, which we 
  // arbitrarily set as 8 (this value can be more or less, 
  // depending on the computing power you wish to utilize)
  hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 8)
  if err != nil {
    return errors.New("unable to encrypt password")
  }
  u.password = hashedPassword
  return nil
}

func (u *User) SetHashedPassword(hashedPassword string) {
  u.password = []byte(hashedPassword)
}

func (u *User) GetPassword() string {
  return string(u.password)
}

func (u *User) CheckPassword(presented string) bool {
  err := bcrypt.CompareHashAndPassword(u.password, []byte(presented))
  return err == nil
}

func (u *User) Save() error {
  return u.persist.UserSave(u)
}

/*
===============================================================================
 Users
-------------------------------------------------------------------------------
 Simple class to group users and help clients easily find them by email
 address or username.
-----------------------------------------------------------------------------*/
type Users []*User

func (list Users) IndexOf(u *User) int {
  idToFind := u.GetId()
  i := 0
  for _, curr := range list {
    if idToFind == curr.GetId() {
      return i
    }
    i++
  }
  return -1  
}

func (list Users) FindByEmail(email string) *User {
   // fmt.Printf("Users.FindByEmail(): listlen = %v\n", len(list))
   for _, curr := range list {
      // fmt.Printf("Users.FindByEmail(): curr = %s\n", curr)
      if email == curr.GetEmail() {
         return curr
      }
   }
   return nil
}

func (list Users) FindById(id string) *User {
   // fmt.Printf("Users.FindById(): listlen = %v\n", len(list))
   for _, curr := range list {
      // fmt.Printf("Users.FindByEmail(): curr = %s\n", curr)
      if id == curr.GetId() {
         return curr
      }
   }
   return nil
}

/*
===============================================================================
 UserAuthentication - HTTP Layer
-------------------------------------------------------------------------------
 This section of the file is HTTP aware and makes use of the previous two
 sections of the file to handle HTTP calls, authenticate users as needed and
 return appropriate errors when it happens.  Note that the UserAuthenticator
 is intended to be used as middleware wrapping any REST API call.
-----------------------------------------------------------------------------*/

/*
==============================================================================
 userCheckAuthToken()
------------------------------------------------------------------------------
 Inputs:  w  ResponseWriter - response to write results / errors into
          r  Request        - request holding the cookie with the token
 Returns:    *User          - authenticated user
 Errors: This function returns no errors.  It either writes an error to the
         response or it returns a successfully authenticated user.

 Takes care of the "http" level of authenticating a user request.  It is
 intended this method be invoked on every request to identify and authenticate
 the user.
============================================================================*/
func userCheckAuthToken(w http.ResponseWriter, r *http.Request) *User {

    // obtain the session token from the request cookies
    c, err := r.Cookie("token")
    if err != nil {
        if err == http.ErrNoCookie {
            // if the cookie is not set, return an unauthorized status
            errorResponse(w, pimErr(authNoToken))
            return nil
        }
        // for any other type of error, return a bad request status
        errorResponse(w, pimErr(badRequest))
        return nil
    }

    // get the JWT string from the cookie
    tknStr := c.Value

    // validate the token and set http responses properly
    var username string
    var errCode  PimErrId
    username, errCode = UserValidateAuthToken(tknStr)
    if (err != nil) {
        errorResponse(w, pimErr(errCode)) // this may need to be typed to pimErr
        return nil
    }

    // now look up the user in our user list and make sure it is there
    user := users.FindByEmail(username)
    if user == nil { // note this error is unlikely since user was in valid token
        log.Printf("userCheckAuthToken() - suspicious activity - valid token with invalid user <%s>\n")
        errorResponse(w, pimErr(authFail))
        return nil // not strictly needed, but return here for clarity
    }

    // return the authenticated user object
    return user
}

/*
==============================================================================
 UserSetAuthToken()
------------------------------------------------------------------------------
 Inputs: w        ResponseWriter - response to write results / cookie into
         username string         - username whose password is already ok-ed

 Errors: This function returns no errors.  It either writes an error to the
         response or it sets the cookie and assumes its caller doesn't care.

 Takes care of the "http" level of signing in with an authentication token.
 This function also decides on the length of the token (TBD move into JWT
 code somehow).
============================================================================*/
func UserSetAuthToken(w http.ResponseWriter, username string) {

    // get a JWT token for this user that expires in the time specified
    expirationTime := time.Now().Add(jwtTimeout)
    tokenString, err := UserGetAuthToken(username, expirationTime)
    if err != nil {
        // if there is an error in creating the JWT return an internal server error
        errorResponse(w, pimErr(authErr))        
        return
    }

    // set the client cookie for "token" as the JWT we just generated
    // we also set an expiry time which is the same as the token itself
    http.SetCookie(w, &http.Cookie{
        Name:    "token",
        Value:   tokenString,
        Expires: expirationTime,
    })    
}

/*
===============================================================================
 UserAuthenticator
-------------------------------------------------------------------------------
 This is the methodology by which every API call make use of the function
 above.  This handler is inserted into the call chain in such a way that no
 actual authenticated handler is ever even called unless authentication works
 properly.

PROBLEM: we can use this in the chain of handlers but HOW do we get the user
we authenticated into the next handler???  Do we need a global, or is there
some way to set it into the request object?  We REALLY don't want to mess
with the Handler types which only take ResponseWriter and Request.
-----------------------------------------------------------------------------*/
func UserAuthenticator(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

      // authenticate the user
      user := userCheckAuthToken(w, r)
      if user == nil {
         return
      }

      // help: how to make the user accessible to the next call?
      // they need it to find their data - let's see if this works
      // but now each handler has to lookup the user - i wish I
      // could set the user into the request itself.
      // r.Header.Set("username", user.GetEmail())  
      ctx := context.WithValue(r.Context(), "user", user)

      // call the next handler - presumable the one doing the real work
      // next.ServeHTTP(w, r)
      next.ServeHTTP(w, r.WithContext(ctx))
   })
}

/*
===============================================================================
 UserFromRequest()
-------------------------------------------------------------------------------
 This is a convenience function to make it easy for each handler to pull the
 user off the request after it has been authenticated.  It is here in this
 file because it is the sister function of the UserAuthenticator handler 
 that sets the context.
-----------------------------------------------------------------------------*/
func UserFromRequest(w http.ResponseWriter, r *http.Request) *User {
   user := r.Context().Value("user")
   if user == nil {
      // this should never happen unless we've called this function
      // from a handler that wasn't wrapped properly in the UserAuthenticator
      // but in the off chance it does we depend on the handler to return
      // immediately.
      log.Printf("UserFromRequest() - coding error - route expects user but no auth set\n")
      errorResponse(w, pimErr(authFail))
      return nil
   }

   return user.(*User)
}

