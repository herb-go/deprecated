package tomluser

import (
	"math/rand"
	"time"

	"github.com/herb-go/herbsecurity/authorize/role"
	"github.com/herb-go/user"
)

var defaultUsersHashMode = "sha256"
var saltlength = 8
var saltchars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func getSalt(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result = result + string(saltchars[rand.Intn(len(saltchars))])
	}
	return result
}

type User struct {
	UID      string
	Password string
	HashMode string
	Salt     string
	Accounts []*user.Account
	Banned   bool
	Roles    *role.Roles
}

func (u *User) Clone() *User {
	newuser := NewUser()
	newuser.UID = u.UID
	newuser.HashMode = u.HashMode
	newuser.Salt = u.Salt
	newuser.Accounts = make([]*user.Account, len(newuser.Accounts))
	copy(newuser.Accounts, u.Accounts)
	newuser.Banned = u.Banned
	roles := make(role.Roles, len(*u.Roles))
	newuser.Roles = &roles
	copy(*newuser.Roles, *u.Roles)
	return newuser
}
func (u *User) SetTo(newuser *User) {
	newuser.UID = u.UID
	newuser.Password = u.Password
	newuser.HashMode = u.HashMode
	newuser.Salt = u.Salt
	newuser.Accounts = u.Accounts
	newuser.Banned = u.Banned
	newuser.Roles = u.Roles
}

func (u *User) VerifyPassword(password string) (bool, error) {
	if u.Password == "" {
		return false, nil
	}
	hashed, err := Hash(u.HashMode, password, u)
	if err != nil {
		return false, err
	}
	return hashed == u.Password, nil
}
func (u *User) UpdatePassword(hashmode string, password string) error {
	newuser := u.Clone()
	newuser.HashMode = hashmode
	newuser.Salt = getSalt(saltlength)
	hashed, err := Hash(hashmode, password, newuser)
	if err != nil {
		return err
	}
	newuser.Password = hashed
	newuser.SetTo(u)
	return nil
}
func NewUser() *User {
	return &User{
		Roles: &role.Roles{},
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
