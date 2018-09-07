package db

import (
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	dbHost         = DbHost
	dbPort         = DbPort
	dbUser         = DbUser
	dbPass         = DbPass
	dbName         = DbName
	authDb         = AuthDb
	collectionName = "notify-message"
)

func initDbHandler() (*mongoHandler, error) {
	dbhandler := &mongoHandler{
		host:     dbHost,
		port:     dbPort,
		database: dbName,
		autdb:    authDb,
		username: dbUser,
		password: dbPass,
	}
	err := dbhandler.GetConnection()
	if err != nil {
		log.Printf("Fail to init db session: %s", err.Error())
	}
	return dbhandler, err
}

func TestNewMongoHandlerConnection(t *testing.T) {
	expectedHandler := &mongoHandler{
		host:     dbHost,
		database: dbName,
		autdb:    authDb,
		username: dbUser,
		password: dbPass,
		port:     dbPort,
		maxIdleTimeMS: 5000,
	}
	dbhandler := NewMongoHandler(dbHost, dbName, authDb, dbUser, dbPass, dbPort, 5000)

	if !reflect.DeepEqual(expectedHandler, dbhandler) {
		t.Fatalf("NewMongoHandler fail: expected %v but got %v", expectedHandler, dbhandler)
	}
}
func TestInitMongoConnection(t *testing.T) {
	dbhandler := &mongoHandler{
		host:     dbHost,
		port:     dbPort,
		database: dbName,
		username: dbUser,
		password: dbPass,
	}
	defer dbhandler.CloseConnection()
	err := dbhandler.GetConnection()
	if err != nil {
		t.Fatalf("Error during create db session %v", err)
	}
	if !dbhandler.IsConnecting() {
		t.Error("Connection must be open after got connecting")
	}
}
func TestInitMongoConnectionFail(t *testing.T) {
	dbhandler := &mongoHandler{
		host:     dbHost,
		port:     dbPort,
		database: dbName,
		username: "Wronginf",
		password: dbPass,
	}
	defer dbhandler.CloseConnection()
	err := dbhandler.GetConnection()
	if err == nil {
		t.Fatalf("Connection must fail")
	}
}

func TestInsertItem(t *testing.T) {
	newMessageID := bson.NewObjectId()
	message := map[string]interface{}{
		"content":      "This is test message",
		"_id":          newMessageID,
		"actorID":      1,
		"targetUserID": 1,
	}
	dbhandler := &mongoHandler{
		host:     dbHost,
		port:     dbPort,
		database: dbName,
		username: dbUser,
		password: dbPass,
	}
	defer dbhandler.CloseConnection()
	err := dbhandler.GetConnection()
	if err != nil {
		t.Fatalf("Fail to init db session: %s", err.Error())
	}
	_, err = dbhandler.AddNewItem(collectionName, message)
	if err != nil {
		t.Fatalf("Insert item must not return error")
	}
}

func TestInsertItemDisconnect(t *testing.T) {
	newMessageID := bson.NewObjectId()
	message := map[string]interface{}{
		"content":      "This is test message",
		"_id":          newMessageID,
		"actorID":      1,
		"targetUserID": 1,
	}
	dbhandler := &mongoHandler{
		host:     dbHost,
		port:     dbPort,
		database: dbName,
		username: dbUser,
		password: dbPass,
	}
	defer dbhandler.CloseConnection()
	err := dbhandler.GetConnection()
	if err != nil {
		t.Fatalf("Fail to init db session: %s", err.Error())
	}
	dbhandler.connection.LogoutAll()
	_, err = dbhandler.AddNewItem(collectionName, message)
	if err == nil {
		t.Fatalf("Insert item must return error")
	}
}

func TestInsertItemDontId(t *testing.T) {
	message := map[string]interface{}{
		"code":         "aaaa",
		"userId":       "289",
		"content":      "This is test message",
		"actorID":      1,
		"targetUserID": 1,
	}
	dbhandler := &mongoHandler{
		host:     dbHost,
		port:     dbPort,
		database: dbName,
		username: dbUser,
		password: dbPass,
	}
	defer dbhandler.CloseConnection()
	err := dbhandler.GetConnection()
	if err != nil {
		t.Fatalf("Fail to init db session: %s", err.Error())
	}
	_, err = dbhandler.AddNewItem(collectionName, message)
	if err != nil {
		t.Fatalf("Insert item must not return error")
	}
}

func TestInsertAndFindById(t *testing.T) {
	newMessageID := bson.NewObjectId()
	createdAt := time.Now()
	message := map[string]interface{}{
		"content":      "This is test message",
		"_id":          newMessageID,
		"actorID":      1,
		"targetUserID": 1,
		"createdAt":    createdAt,
	}
	dbhandler := &mongoHandler{
		host:     dbHost,
		port:     dbPort,
		database: dbName,
		autdb:    authDb,
		username: dbUser,
		password: dbPass,
	}
	defer dbhandler.CloseConnection()
	err := dbhandler.GetConnection()
	if err != nil {
		t.Fatalf("Fail to init db session: %s", err.Error())
	}
	_, err = dbhandler.AddNewItem(collectionName, message)
	if err != nil {
		t.Fatalf("Insert item must not return error")
	}
	actualMessage, err := dbhandler.FindItemByID(collectionName, newMessageID)
	if err != nil {
		t.Fatalf("Error during find message by ID: %s", err.Error())
	}
	if message["content"] != actualMessage["content"] {
		t.Fatalf("Found and inserted not match: expected %v but got %v", message, actualMessage)
	}
}

func TestFindAll(t *testing.T) {
	dbhandler, err := initDbHandler()
	defer dbhandler.CloseConnection()
	if err != nil {
		t.Fatalf("Fail when init db")
	}
	results, err := dbhandler.GetAllItems(collectionName, "DESC", "createdAt", 10, 1, map[string]interface{}{"actorID": 1})
	if err != nil {
		t.Fatalf("Error when get all items %s", err.Error())
	}
	if results.PageSize > 10 {
		t.Fatalf("Number of returned items cannot be greater than limit")
	}
}

func TestGetAllItemsNoLimit(t *testing.T) {
	dbhandler, err := initDbHandler()
	defer dbhandler.CloseConnection()
	if err != nil {
		t.Fatalf("Fail when init db")
	}
	_, err = dbhandler.GetAllItemsNoLimit(collectionName, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Error when get all items %s", err.Error())
	}
}

func TestGetTotalNotifyMessage(t *testing.T) {
	dbhandler, err := initDbHandler()
	defer dbhandler.CloseConnection()
	if err != nil {
		t.Fatalf("Fail when init db")
	}
	_, err = dbhandler.GetTotal(collectionName, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Error when get all items %s", err.Error())
	}
}

func TestRemoveItemByID(t *testing.T) {
	dbhandler, err := initDbHandler()
	defer dbhandler.CloseConnection()
	newMessageID := bson.NewObjectId()
	createdAt := time.Now()
	message := map[string]interface{}{
		"content":      "This is test message",
		"_id":          newMessageID,
		"actorID":      1,
		"targetUserID": 1,
		"createdAt":    createdAt,
	}
	_, err = dbhandler.AddNewItem(collectionName, message)
	if err != nil {
		t.Fatalf("Insert item must not return error")
	}
	find, err := dbhandler.FindItemByID(collectionName, newMessageID)
	if err != nil {
		t.Fatalf("Error during find message by ID: %s", err.Error())
	}
	stringID, _ := newMessageID.MarshalText()
	dbhandler.RemoveItemByID(dbName, find["_id"].(string))
	_, err = dbhandler.FindItemByID(dbName, string(stringID))
	if err == nil {
		t.Fatalf("After deleting, find must return error")
	}
}

func TestRemoveItemBy(t *testing.T) {
	dbhandler, err := initDbHandler()
	defer dbhandler.CloseConnection()
	newMessageID := bson.NewObjectId()
	createdAt := time.Now()
	message := map[string]interface{}{
		"content":      "This is test message",
		"_id":          newMessageID,
		"actorID":      1,
		"targetUserID": 1,
		"createdAt":    createdAt,
	}
	_, err = dbhandler.AddNewItem(collectionName, message)
	if err != nil {
		t.Fatalf("Insert item must not return error")
	}

	filter := map[string]interface{}{
		"actorID": 1,
	}
	_, err = dbhandler.FindBy(collectionName, filter)
	if err != nil {
		t.Fatalf("Error during find message by ID: %s", err.Error())
	}
	stringID, _ := newMessageID.MarshalText()
	dbhandler.RemoveItemBy(dbName, filter)
	_, err = dbhandler.FindItemByID(dbName, string(stringID))
	if err == nil {
		t.Fatalf("After deleting, find must return error")
	}
}

func TestInvalidRemoveItemByID(t *testing.T) {
	dbhandler, err := initDbHandler()
	defer dbhandler.CloseConnection()
	newMessageID := bson.NewObjectId()
	createdAt := time.Now()
	message := map[string]interface{}{
		"content":      "This is test message",
		"_id":          newMessageID,
		"actorID":      1,
		"targetUserID": 1,
		"createdAt":    createdAt,
	}
	_, err = dbhandler.AddNewItem(collectionName, message)
	if err != nil {
		t.Fatalf("Insert item must not return error")
	}
	errRemove := dbhandler.RemoveItemByID(dbName, "fdsafas")
	if errRemove == nil {
		t.Fatalf("Remove must not return error")
	}
}

func TestInvalidFindID(t *testing.T) {
	dbhandler, err := initDbHandler()
	defer dbhandler.CloseConnection()
	newMessageID := bson.NewObjectId()
	createdAt := time.Now()
	message := map[string]interface{}{
		"content":      "This is test message",
		"_id":          newMessageID,
		"actorID":      1,
		"targetUserID": 1,
		"createdAt":    createdAt,
	}
	_, err = dbhandler.AddNewItem(collectionName, message)
	if err != nil {
		t.Fatalf("Insert item must not return error")
	}
	find, err := dbhandler.FindItemByID(collectionName, "")
	if err == nil {
		t.Fatalf("Find id must be return error: %s", err.Error())
		t.Fatalf("result: %+v", find)
	}
	dbhandler.RemoveItemByID(collectionName, string(newMessageID))
}

func TestUpdateBy(t *testing.T) {
	dbhandler, err := initDbHandler()
	defer dbhandler.CloseConnection()
	newMessageID := bson.NewObjectId()
	createdAt := time.Now()
	message := map[string]interface{}{
		"content":      "This is test message",
		"_id":          newMessageID,
		"actorID":      2,
		"targetUserID": 12,
		"createdAt":    createdAt,
		"seen":         false,
	}
	_, err = dbhandler.AddNewItem(collectionName, message)
	if err != nil {
		t.Fatalf("Error during save message by ID: %s", err.Error())
	}
	insertedItem, err := dbhandler.FindItemByID(collectionName, newMessageID)
	if err != nil {
		t.Fatalf("Error during find message by ID: %s", err.Error())
	}
	insertedItem["seen"] = !(insertedItem["seen"]).(bool)
	clonedItem := cloneStringMap(insertedItem)
	selector := map[string]interface{}{
		"targetUserID": 12,
	}
	update := map[string]interface{}{
		"seen": false,
	}
	_, err = dbhandler.UpdateBy(collectionName, selector, update)
	if err != nil {
		t.Fatalf("Update by id must not return error but got %s", err.Error())
	}
	if !reflect.DeepEqual(insertedItem, clonedItem) {
		t.Fatalf("Update must not modify original item")
	}
	dbhandler.RemoveItemByID(collectionName, insertedItem["_id"].(string))
}

func TestUpdateByID(t *testing.T) {
	dbhandler, err := initDbHandler()
	defer dbhandler.CloseConnection()
	newMessageID := bson.NewObjectId()
	createdAt := time.Now()
	message := map[string]interface{}{
		"content":      "This is test message",
		"_id":          newMessageID,
		"actorID":      2,
		"targetUserID": 12,
		"createdAt":    createdAt,
		"seen":         false,
	}
	_, err = dbhandler.AddNewItem(collectionName, message)
	if err != nil {
		t.Fatalf("Error during save message by ID: %s", err.Error())
	}
	insertedItem, err := dbhandler.FindItemByID(collectionName, newMessageID)
	if err != nil {
		t.Fatalf("Error during find message by ID: %s", err.Error())
	}
	insertedItem["seen"] = !(insertedItem["seen"]).(bool)
	clonedItem := cloneStringMap(insertedItem)
	err = dbhandler.UpdateByID(collectionName, (insertedItem["_id"].(string)), insertedItem)
	if err != nil {
		t.Fatalf("Update by id must not return error but got %s", err.Error())
	}
	if !reflect.DeepEqual(insertedItem, clonedItem) {
		t.Fatalf("Update must not modify original item")
	}
	dbhandler.RemoveItemByID(collectionName, insertedItem["_id"].(string))
}

func TestUpdateByIDDisconnect(t *testing.T) {
	dbhandler, err := initDbHandler()
	defer dbhandler.CloseConnection()
	newMessageID := bson.NewObjectId()
	createdAt := time.Now()
	message := map[string]interface{}{
		"content":      "This is test message",
		"_id":          newMessageID,
		"actorID":      2,
		"targetUserID": 12,
		"createdAt":    createdAt,
		"seen":         false,
	}
	_, err = dbhandler.AddNewItem(collectionName, message)
	if err != nil {
		t.Fatalf("Error during save message by ID: %s", err.Error())
	}
	insertedItem, err := dbhandler.FindItemByID(collectionName, newMessageID)
	if err != nil {
		t.Fatalf("Error during find message by ID: %s", err.Error())
	}
	dbhandler.connection.LogoutAll()
	err = dbhandler.UpdateByID(collectionName, (insertedItem["_id"].(string)), insertedItem)
	if err == nil {
		t.Fatalf("Update by id must return error but got %s", err.Error())
	}
}

func TestInvalidUpdateByID(t *testing.T) {
	dbhandler, err := initDbHandler()
	defer dbhandler.CloseConnection()
	newMessageID := bson.NewObjectId()
	createdAt := time.Now()
	message := map[string]interface{}{
		"content":      "This is test message",
		"_id":          newMessageID,
		"actorID":      2,
		"targetUserID": 12,
		"createdAt":    createdAt,
		"seen":         false,
	}
	_, err = dbhandler.AddNewItem(collectionName, message)
	if err != nil {
		t.Fatalf("Error during save message by ID: %s", err.Error())
	}
	insertedItem, err := dbhandler.FindItemByID(collectionName, newMessageID)
	if err != nil {
		t.Fatalf("Error during find message by ID: %s", err.Error())
	}
	insertedItem["seen"] = !(insertedItem["seen"]).(bool)
	err = dbhandler.UpdateByID(collectionName, "", insertedItem)
	if err == nil {
		t.Fatalf("Update by id must return error but got %s", err.Error())
	}
}

func TestInvalidObjectIDError_Error(t *testing.T) {
	type fields struct {
		message string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := InvalidObjectIDError{
				message: tt.fields.message,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("InvalidObjectIDError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_initDbHandler(t *testing.T) {
	tests := []struct {
		name    string
		want    *mongoHandler
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := initDbHandler()
			if (err != nil) != tt.wantErr {
				t.Errorf("initDbHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initDbHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mongoHandler_AddNewItem(t *testing.T) {
	type fields struct {
		host       string
		port       int
		database   string
		autdb      string
		username   string
		password   string
		connection *mgo.Session
	}
	type args struct {
		dataName string
		item     map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoHandler{
				host:       tt.fields.host,
				port:       tt.fields.port,
				database:   tt.fields.database,
				autdb:      tt.fields.autdb,
				username:   tt.fields.username,
				password:   tt.fields.password,
				connection: tt.fields.connection,
			}
			got, err := m.AddNewItem(tt.args.dataName, tt.args.item)
			if (err != nil) != tt.wantErr {
				t.Errorf("mongoHandler.AddNewItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mongoHandler.AddNewItem() = %v, want %v", got, tt.want)
			}
		})
	}
}
