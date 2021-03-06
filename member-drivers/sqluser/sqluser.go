package sqluser

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"crypto/rand"
	"crypto/sha256"

	"github.com/herb-go/datasource/sql/db"
	"github.com/herb-go/datasource/sql/querybuilder"
	"github.com/herb-go/datasource/sql/querybuilder/modelmapper"
	"github.com/herb-go/user"
	"github.com/herb-go/deprecated/member"
)

const (
	//FlagEmpty sql user create flag empty
	FlagEmpty = 0
	//FlagWithAccount sql user create flag with account module
	FlagWithAccount = 1
	//FlagWithPassword sql user create flag with password module
	FlagWithPassword = 2
	//FlagWithToken sql user create flag with token module
	FlagWithToken = 4
	//FlagWithUser sql user create flag with user module
	FlagWithUser = 8
)

//RandomBytesLength bytes length for RandomBytes function.
var RandomBytesLength = 32

//ErrHashMethodNotFound error raised when password hash method not found.
var ErrHashMethodNotFound = errors.New("password hash method not found")

//HashFunc interaface of pasword hash func
type HashFunc func(key string, salt string, password string) ([]byte, error)

//DefaultAccountMapperName default database table name for module account.
var DefaultAccountMapperName = "account"

//DefaultPasswordMapperName default database table name for module password.
var DefaultPasswordMapperName = "password"

//DefaultTokenMapperName default database table name for module token.
var DefaultTokenMapperName = "token"

//DefaultUserMapperName default database table name for module user.
var DefaultUserMapperName = "user"

//DefaultHashMethod default hash method when created password data.
var DefaultHashMethod = "sha256"

//HashFuncMap all available password hash func.
//You can insert custom hash func into this map.
var HashFuncMap = map[string]HashFunc{
	"sha256": func(key string, salt string, password string) ([]byte, error) {
		var val = []byte(key + salt + password)
		var s256 = sha256.New()
		s256.Write(val)
		val = s256.Sum(nil)
		s256.Write(val)
		return []byte(hex.EncodeToString(s256.Sum(nil))), nil
	},
}

//New create User framework with given database ,uidgeneraterand falg.
//flag is values combine with flags to special which modules used.
//For example ,New(db,FlagWithAccount | FlagWithToken)
func New(db db.Database, uidgenerater func() (string, error), flag int) *User {
	q := querybuilder.New()
	q.Driver = db.Driver()
	return &User{
		DB: db,
		Tables: Tables{
			AccountMapperName:  DefaultAccountMapperName,
			PasswordMapperName: DefaultPasswordMapperName,
			TokenMapperName:    DefaultTokenMapperName,
			UserMapperName:     DefaultUserMapperName,
		},
		HashMethod:     DefaultHashMethod,
		UIDGenerater:   uidgenerater,
		TokenGenerater: Timestamp,
		SaltGenerater:  RandomBytes,
		Flag:           flag,
		QueryBuilder:   q,
	}
}

//Tables struct stores table info.
type Tables struct {
	AccountMapperName  string
	PasswordMapperName string
	TokenMapperName    string
	UserMapperName     string
}

//RandomBytes string generater return random bytes.
//Default length is 32 byte.You can change default length by change sqluesr.RandomBytesLength .
func RandomBytes() (string, error) {
	var bytes = make([]byte, RandomBytesLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

//Timestamp string generater return timestamp in nano.
func Timestamp() (string, error) {
	return strconv.FormatInt(time.Now().UnixNano(), 10), nil
}

//User main struct of sqluser module.
type User struct {
	//DB database used.
	DB db.Database
	//Tables table name info.
	Tables Tables
	//Flag sqluser modules create flg.
	Flag int
	//UIDGenerater string generater for uid
	//default value is uuid
	UIDGenerater func() (string, error)
	//TokenGenerater string generater for usertoken
	//default value is timestamp
	TokenGenerater func() (string, error)
	//SaltGenerater string generater for salt
	//default value is 32 byte length random bytes.
	SaltGenerater func() (string, error)
	//HashMethod hash method which used to generate new salt.
	//default value is sha256
	HashMethod string
	//PasswordKey static key used in passwrod hash generater.
	//default value is empty.
	//You can change this value after sqluser init.
	PasswordKey string
	//QueryBuilder sql query builder
	QueryBuilder *querybuilder.Builder
}

//AddTablePrefix add prefix to user table names.
func (u *User) AddTablePrefix(prefix string) {
	u.Tables.AccountMapperName = prefix + u.Tables.AccountMapperName
	u.Tables.PasswordMapperName = prefix + u.Tables.PasswordMapperName
	u.Tables.TokenMapperName = prefix + u.Tables.TokenMapperName
	u.Tables.UserMapperName = prefix + u.Tables.UserMapperName
}

//HasFlag check if sqluser module created with special flag.
func (u *User) HasFlag(flag int) bool {
	return u.Flag&flag != 0
}

//AccountTableName return actual account database table name.
func (u *User) AccountTableName() string {
	return u.DB.BuildTableName(u.Tables.AccountMapperName)
}

//PasswordTableName return actual password database table name.
func (u *User) PasswordTableName() string {
	return u.DB.BuildTableName(u.Tables.PasswordMapperName)
}

//TokenTableName return actual token database table name.
func (u *User) TokenTableName() string {
	return u.DB.BuildTableName(u.Tables.TokenMapperName)
}

//UserTableName return actual user database table name.
func (u *User) UserTableName() string {
	return u.DB.BuildTableName(u.Tables.UserMapperName)
}

//Account return account mapper
func (u *User) Account() *AccountMapper {
	return &AccountMapper{
		ModelMapper: modelmapper.New(db.NewTable(u.DB, u.Tables.AccountMapperName)),
		User:        u,
	}
}

//Password return password mapper
func (u *User) Password() *PasswordMapper {
	return &PasswordMapper{
		ModelMapper: modelmapper.New(db.NewTable(u.DB, u.Tables.PasswordMapperName)),
		User:        u,
	}
}

//Token return token mapper
func (u *User) Token() *TokenMapper {
	return &TokenMapper{
		ModelMapper: modelmapper.New(db.NewTable(u.DB, u.Tables.TokenMapperName)),
		User:        u,
	}
}

//User return user mapper
func (u *User) User() *UserMapper {
	return &UserMapper{
		ModelMapper: modelmapper.New(db.NewTable(u.DB, u.Tables.UserMapperName)),
		User:        u,
	}
}

//AccountMapper account mapper
type AccountMapper struct {
	*modelmapper.ModelMapper
	User    *User
	Service *member.Service
}

//Execute install account module to member service as provider
func (a *AccountMapper) Execute(service *member.Service) {
	service.AccountsProvider = a
	a.Service = service
}

//Unbind unbind account from user.
//Return any error if raised.
func (a *AccountMapper) Unbind(uid string, account *user.Account) error {
	query := a.User.QueryBuilder
	tx, err := a.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	Delete := query.NewDeleteQuery(a.TableName())
	Delete.Where.Condition = query.And(
		query.Equal("account.uid", uid),
		query.Equal("account.keyword", account.Keyword),
		query.Equal("account.account", account.Account),
	)
	_, err = Delete.Query().Exec(tx)
	if err != nil {
		return err
	}
	return tx.Commit()

}

//Bind bind account to user.
//Return any error if raised.
//If account exists, error user.ErrAccountBindingExists will raised.
func (a *AccountMapper) Bind(uid string, account *user.Account) error {
	query := a.User.QueryBuilder
	tx, err := a.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var u = ""
	Select := query.NewSelectQuery()
	Select.Select.Add("account.uid")
	Select.From.AddAlias("account", a.TableName())
	Select.Where.Condition = query.And(
		query.Equal("keyword", account.Keyword),
		query.Equal("account", account.Account),
	)
	row := Select.QueryRow(a.DB())
	err = row.Scan(&u)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	} else {
		return user.ErrAccountBindingExists

	}

	var CreatedTime = time.Now().Unix()
	Insert := query.NewInsertQuery(a.TableName())
	Insert.Insert.
		Add("uid", uid).
		Add("keyword", account.Keyword).
		Add("account", account.Account).
		Add("created_time", CreatedTime)
	_, err = Insert.Query().Exec(tx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

//FindOrInsert find user by account.if account did not exists,a new user with given account will be created.
//UIDGenerater used when create new user.
//Return user id and any error if raised.
func (a *AccountMapper) FindOrInsert(UIDGenerater func() (string, error), account *user.Account) (string, bool, error) {
	query := a.User.QueryBuilder
	var result = AccountModel{}
	tx, err := a.DB().Begin()
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback()
	Select := query.NewSelectQuery()
	Select.From.AddAlias("account", a.TableName())
	Select.Select.Add("account.uid", "account.keyword", "account.account", "account.created_time")
	Select.Where.Condition = query.And(
		query.Equal("account.keyword", account.Keyword),
		query.Equal("account.account", account.Account),
	)
	row := Select.QueryRow(a.DB())
	err = Select.Result().
		Bind("account.uid", &result.UID).
		Bind("account.keyword", &result.Keyword).
		Bind("account.account", &result.Account).
		Bind("account.created_time", &result.CreatedTime).
		ScanFrom(row)
	if err == nil {
		return result.UID, false, nil
	}
	if err != sql.ErrNoRows {
		return "", false, err
	}
	uid, err := UIDGenerater()
	var CreatedTime = time.Now().Unix()
	Insert := query.NewInsertQuery(a.TableName())
	Insert.Insert.
		Add("uid", uid).
		Add("keyword", account.Keyword).
		Add("account", account.Account).
		Add("created_time", CreatedTime)
	_, err = Insert.Query().Exec(tx)
	if err != nil {
		return "", false, err
	}
	if a.User.HasFlag(FlagWithUser) {
		Insert := query.NewInsertQuery(a.User.UserTableName())
		Insert.Insert.
			Add("uid", uid).
			Add("status", member.StatusNormal).
			Add("created_time", CreatedTime).
			Add("updated_time", CreatedTime)
		_, err = Insert.Query().Exec(tx)
		if err != nil {
			return "", false, err
		}
	}
	return uid, true, tx.Commit()
}

//Insert create new user with given account.
//Return any error if raised.
//If account exists,member.ErrAccountRegisterExists will raise.
func (a *AccountMapper) Insert(uid string, keyword string, account string) error {
	query := a.User.QueryBuilder
	tx, err := a.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var u = ""
	Select := query.NewSelectQuery()
	Select.Select.Add("uid")
	Select.From.Add(a.TableName())
	Select.Where.Condition = query.And(
		query.Equal("keyword", keyword),
		query.Equal("account", account),
	)
	row := Select.QueryRow(a.DB())
	err = row.Scan(&u)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	} else {
		return member.ErrAccountRegisterExists
	}
	var CreatedTime = time.Now().Unix()
	Insert := query.NewInsertQuery(a.TableName())
	Insert.Insert.
		Add("uid", uid).
		Add("keyword", keyword).
		Add("account", account).
		Add("created_time", CreatedTime)
	_, err = Insert.Query().Exec(tx)
	if err != nil {
		return err
	}
	if a.User.HasFlag(FlagWithUser) {
		Insert := query.NewInsertQuery(a.User.UserTableName())
		Insert.Insert.
			Add("uid", uid).
			Add("status", member.StatusNormal).
			Add("created_time", CreatedTime).
			Add("updated_time", CreatedTime)
		_, err = Insert.Query().Exec(tx)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

//Find find account by given keyword and account.
//Return account model and any error if raised.
func (a *AccountMapper) Find(keyword string, account string) (AccountModel, error) {
	query := a.User.QueryBuilder
	var result = AccountModel{}
	if keyword == "" || account == "" {
		return result, sql.ErrNoRows
	}
	Select := query.NewSelectQuery()
	Select.Select.Add("uid", "keyword", "account", "created_time")
	Select.From.Add(a.TableName())
	Select.Where.Condition = query.And(
		query.Equal("keyword", keyword),
		query.Equal("account", account),
	)
	row := Select.QueryRow(a.DB())
	err := Select.Result().
		Bind("uid", &result.UID).
		Bind("keyword", &result.Keyword).
		Bind("account", &result.Account).
		Bind("created_time", &result.CreatedTime).
		ScanFrom(row)
	return result, err
}

//FindAllByUID find account models by user id list.
//Retrun account models and any error if rased.
func (a *AccountMapper) FindAllByUID(uids ...string) ([]AccountModel, error) {
	query := a.User.QueryBuilder
	var result = []AccountModel{}
	if len(uids) == 0 {
		return result, nil
	}
	Select := query.NewSelectQuery()
	Select.Select.Add("account.uid", "account.keyword", "account.account")
	Select.From.AddAlias("account", a.TableName())
	Select.Where.Condition = query.In("account.uid", uids)
	rows, err := Select.QueryRows(a.DB())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		v := AccountModel{}
		err := Select.Result().
			Bind("account.uid", &v.UID).
			Bind("account.keyword", &v.Keyword).
			Bind("account.account", &v.Account).
			ScanFrom(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

//Accounts get member account map by user id list.
//Return account map and any error if rasied.
//User unfound in account map will be a nil value.
func (a *AccountMapper) Accounts(uid ...string) (*member.Accounts, error) {
	models, err := a.FindAllByUID(uid...)
	if err != nil {
		return nil, err
	}
	result := member.Accounts{}
	for _, v := range models {
		if result[v.UID] == nil {
			result[v.UID] = user.Accounts{}
		}
		account := user.Account{Keyword: v.Keyword, Account: v.Account}
		result[v.UID] = append(result[v.UID], &account)
	}
	return &result, nil
}

//AccountToUID find user by account.
//Return user id and any error if rasied.
//If user not found,a empty string will be returned.
func (a *AccountMapper) AccountToUID(account *user.Account) (uid string, err error) {
	model, err := a.Find(account.Keyword, account.Account)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return model.UID, err
}

//Register register a user with special account.
//Return user id and any error if raised.
//If account exists,member.ErrAccountRegisterExists will raise.
func (a *AccountMapper) Register(account *user.Account) (uid string, err error) {
	uid, err = a.User.UIDGenerater()
	if err != nil {
		return
	}
	err = a.Insert(uid, account.Keyword, account.Account)
	return
}

//AccountToUIDOrRegister find a user by account.if user didnot exist,a new user will be created.
//Return user id and any error if raised.
func (a *AccountMapper) AccountToUIDOrRegister(account *user.Account) (uid string, registerd bool, err error) {
	return a.FindOrInsert(a.User.UIDGenerater, account)
}

//BindAccount bind account to user.
//Return any error if rasied.
//If account exists, error user.ErrAccountBindingExists will raised.
func (a *AccountMapper) BindAccount(uid string, account *user.Account) error {
	return a.Bind(uid, account)
}

//UnbindAccount unbind account from user.
//Return any error if rasied.
func (a *AccountMapper) UnbindAccount(uid string, account *user.Account) error {
	return a.Unbind(uid, account)
}

//AccountModel account data model
type AccountModel struct {
	//UID user id.
	UID string
	//Keyword account keyword.
	Keyword string
	//Account account name.
	Account string
	//CreatedTime created timestamp in second.
	CreatedTime int64
}

//PasswordMapper password mapper
type PasswordMapper struct {
	*modelmapper.ModelMapper
	User    *User
	Service *member.Service
}

//Execute install passowrd module to member service as provider
func (p *PasswordMapper) Execute(service *member.Service) {
	service.PasswordProvider = p
	p.Service = service
}

//PasswordChangeable return password changeable
func (p *PasswordMapper) PasswordChangeable() bool {
	return true
}

//Find find password model by userd id.
//Return any error if raised.
func (p *PasswordMapper) Find(uid string) (PasswordModel, error) {
	query := p.User.QueryBuilder
	var result = PasswordModel{}
	if uid == "" {
		return result, sql.ErrNoRows
	}
	Select := query.NewSelectQuery()
	Select.Select.Add("password.hash_method", "password.salt", "password.password", "password.updated_time")
	Select.From.AddAlias("password", p.TableName())
	Select.Where.Condition = query.Equal("uid", uid)
	q := Select.Query()
	row := p.DB().QueryRow(q.QueryCommand(), q.QueryArgs()...)
	result.UID = uid
	args := Select.Result().
		Bind("password.hash_method", &result.HashMethod).
		Bind("password.salt", &result.Salt).
		Bind("password.password", &result.Password).
		Bind("password.updated_time", &result.UpdatedTime).
		Pointers()

	err := row.Scan(args...)
	return result, err
}

//InsertOrUpdate insert or update password model.
//Return any error if raised.
func (p *PasswordMapper) InsertOrUpdate(model *PasswordModel) error {
	query := p.User.QueryBuilder

	tx, err := p.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	Update := query.NewUpdateQuery(p.TableName())
	Update.Update.
		Add("hash_method", model.HashMethod).
		Add("salt", model.Salt).
		Add("password", model.Password).
		Add("updated_time", model.UpdatedTime)
	Update.Where.Condition = query.Equal("uid", model.UID)
	r, err := Update.Query().Exec(tx)

	if err != nil {
		return err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 0 {
		return tx.Commit()
	}
	Insert := query.NewInsertQuery(p.TableName())
	Insert.Insert.
		Add("uid", model.UID).
		Add("hash_method", model.HashMethod).
		Add("salt", model.Salt).
		Add("password", model.Password).
		Add("updated_time", model.UpdatedTime)
	_, err = Insert.Query().Exec(tx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

//VerifyPassword Verify user password.
//Return verify and any error if raised.
//if user not found,error member.ErrUserNotFound will be raised.
func (p *PasswordMapper) VerifyPassword(uid string, password string) (bool, error) {
	model, err := p.Find(uid)
	if err == sql.ErrNoRows {
		return false, member.ErrUserNotFound
	}
	if err != nil {
		return false, err
	}
	hash := HashFuncMap[model.HashMethod]
	if hash == nil {
		return false, ErrHashMethodNotFound
	}
	hashed, err := hash(p.User.PasswordKey, model.Salt, password)
	if err != nil {
		return false, err
	}
	return bytes.Compare(hashed, model.Password) == 0, nil
}

//UpdatePassword update user password.If user password does not exist,new password record will be created.
//Return any error if raised.
func (p *PasswordMapper) UpdatePassword(uid string, password string) error {
	salt, err := p.User.SaltGenerater()
	if err != nil {
		return err
	}
	hash := HashFuncMap[p.User.HashMethod]
	if hash == nil {
		return ErrHashMethodNotFound
	}
	hashed, err := hash(p.User.PasswordKey, salt, password)
	if err != nil {
		return err
	}
	model := &PasswordModel{
		UID:         uid,
		HashMethod:  p.User.HashMethod,
		Salt:        salt,
		Password:    hashed,
		UpdatedTime: time.Now().Unix(),
	}
	return p.InsertOrUpdate(model)
}

//PasswordModel password data model
type PasswordModel struct {
	//UID user id.
	UID string
	//HashMethod hash method to verify this password.
	HashMethod string
	//Salt random salt.
	Salt string
	//Password hashed password data.
	Password []byte
	//UpdatedTime updated timestamp in second.
	UpdatedTime int64
}

//TokenMapper token mapper
type TokenMapper struct {
	*modelmapper.ModelMapper
	User    *User
	Service *member.Service
}

//Execute install token module to member service as provider
func (t *TokenMapper) Execute(service *member.Service) {
	service.TokenProvider = t
	t.Service = service
}

//InsertOrUpdate insert or update user token record.
func (t *TokenMapper) InsertOrUpdate(uid string, token string) error {
	query := t.User.QueryBuilder

	tx, err := t.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var CreatedTime = time.Now().Unix()
	Update := query.NewUpdateQuery(t.TableName())
	Update.Update.
		Add("token", token).
		Add("updated_time", CreatedTime)
	Update.Where.Condition = query.Equal("uid", uid)
	r, err := Update.Query().Exec(tx)
	if err != nil {
		return err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 0 {
		return tx.Commit()
	}
	Insert := query.NewInsertQuery(t.TableName())
	Insert.Insert.
		Add("uid", uid).
		Add("token", token).
		Add("updated_time", CreatedTime)
	_, err = Insert.Query().Exec(tx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

//FindAllByUID find all token model by uid list.
//Return token models and any error if raised.
func (t *TokenMapper) FindAllByUID(uids ...string) ([]TokenModel, error) {
	query := t.User.QueryBuilder
	var result = []TokenModel{}
	if len(uids) == 0 {
		return result, nil
	}
	Select := query.NewSelectQuery()
	Select.Select.Add("token.uid", "token.token")
	Select.From.AddAlias("token", t.TableName())
	Select.Where.Condition = query.In("token.uid", uids)
	rows, err := Select.QueryRows(t.DB())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		v := TokenModel{}
		err = rows.Scan(&v.UID, &v.Token)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

//Tokens get member token map by user id list.
//Return token map and any error if rasied.
//User unfound in token map will be a nil value.
func (t *TokenMapper) Tokens(uid ...string) (member.Tokens, error) {
	models, err := t.FindAllByUID(uid...)
	if err != nil {
		return nil, err
	}
	result := member.Tokens{}
	for _, v := range models {
		result[v.UID] = v.Token
	}
	return result, nil

}

//Revoke revoke and regenerate a new token to user.if revoke record does not exist,a new record will be created.
//Return new user token and any error if raised.
func (t *TokenMapper) Revoke(uid string) (string, error) {
	token, err := t.User.TokenGenerater()
	if err != nil {
		return "", err
	}
	return token, t.InsertOrUpdate(uid, token)
}

//TokenModel token data model
type TokenModel struct {
	//UID user id
	UID string
	//Token current user token
	Token string
	//UpdatedTime updated timestamp in second.
	UpdatedTime string
}

//UserMapper user mapper
type UserMapper struct {
	*modelmapper.ModelMapper
	User    *User
	Service *member.Service
}

//Execute install user module to member service as provider
func (u *UserMapper) Execute(service *member.Service) {
	service.StatusProvider = u
	u.Service = service
}

//FindAllByUID find user models by user id list.
//Return User model list and any error if raised.
func (u *UserMapper) FindAllByUID(uids ...string) ([]UserModel, error) {
	query := u.User.QueryBuilder

	var result = []UserModel{}
	if len(uids) == 0 {
		return result, nil
	}
	Select := query.NewSelectQuery()
	Select.Select.Add("user.uid", "user.status")
	Select.From.AddAlias("user", u.TableName())
	Select.Where.Condition = query.In("user.uid", uids)
	rows, err := Select.QueryRows(u.DB())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		v := UserModel{}
		err = rows.Scan(&v.UID, &v.Status)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

//InsertOrUpdate insert or update user model with status.
//Return any error if raised.
func (u *UserMapper) InsertOrUpdate(uid string, status member.Status) error {
	query := u.User.QueryBuilder
	tx, err := u.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var CreatedTime = time.Now().Unix()
	Update := query.NewUpdateQuery(u.TableName())
	Update.Update.
		Add("status", status).
		Add("updated_time", CreatedTime)
	Update.Where.Condition = query.Equal("uid", uid)
	r, err := Update.Query().Exec(tx)
	if err != nil {
		return err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 0 {
		return tx.Commit()
	}
	Insert := query.NewInsertQuery(u.TableName())
	Insert.Insert.
		Add("uid", uid).
		Add("status", status).
		Add("updated_time", CreatedTime).
		Add("created_time", CreatedTime)
	_, err = Insert.Query().Exec(tx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

//Statuses get member  status map by user id list.
//Return  status map and any error if rasied.
//User unfound in token map will be false.
func (u *UserMapper) Statuses(uid ...string) (member.StatusMap, error) {
	models, err := u.FindAllByUID(uid...)
	if err != nil {
		return nil, err
	}
	result := member.StatusMap{}
	for _, v := range models {
		result[v.UID] = member.Status(v.Status)
	}
	return result, nil
}

//SupportedStatus return supported status map
func (u *UserMapper) SupportedStatus() map[member.Status]bool {
	return member.StatusMapAll
}

//SetStatus set user  status.
//Return any error if raised.
func (u *UserMapper) SetStatus(uid string, status member.Status) error {
	return u.InsertOrUpdate(uid, status)
}

//UserModel user data model
type UserModel struct {
	//UID user id
	UID string
	//CreatedTime created timestamp in second
	CreatedTime int64
	//UpdateTIme updated timestamp in second
	UpdateTIme int64
	//Status user status
	Status int
}
