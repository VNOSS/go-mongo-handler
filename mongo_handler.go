package db

import (
	"errors"
	"log"
	"notify-message/helper"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// InvalidObjectIDError is returned when wrong object id passed
type InvalidObjectIDError struct {
	message string
}

func (e InvalidObjectIDError) Error() string {
	return e.message
}

type mongoHandler struct {
	host       string
	port       int
	database   string
	autdb      string
	username   string
	password   string
	maxIdleTimeMS int
	connection *mgo.Session
}

func (m *mongoHandler) GetConnection() error {
	if m.connection == nil {
		var err error
		m.connection, err = m.createMongoSession()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoHandler) IsConnecting() bool {
	return m.connection != nil
}

func (m *mongoHandler) CloseConnection() {
	if m.connection != nil {
		m.connection.Close()
		m.connection = nil
	}
}

// GetAllItems get all items with paging infor
func (m *mongoHandler) GetAllItems(dataname, orderBy, sortBy string, limit, page int, filters map[string]interface{}) (PagedResults, error) {
	// Make sure connection open
	err := m.GetConnection()
	if err != nil {
		log.Printf("[App.db]: Error during create mongo session: %s\n", err)
		return PagedResults{}, err
	}
	workingDBSession := m.connection.Copy()
	defer workingDBSession.Close()
	//var cursorFields  []string
	c := workingDBSession.DB(m.database).C(dataname)
	// Get total items by filters
	total, err := c.Find(filters).Count()
	if err != nil {
		log.Printf("[App.db]: Error during couting items: %s\n", err)
		return PagedResults{}, err
	}
	pagingInfor := helper.NewPaginator(total, limit, page)
	// Create sortby string
	sortString := "+" + sortBy
	if strings.ToUpper(orderBy) == "DESC" {
		sortString = "-" + sortBy
	}
	// First we need to skip previous page items
	skip := (page * limit) - limit
	//q := minquery.New(workingDBSession.DB(m.database), dataname, filters).Sort(sortString).Limit(skip)
	q := c.Find(filters).Sort(sortString).Skip(skip)
	var items []interface{}
	err = q.Limit(limit).All(&items)
	// This will move the cursort to the last item need to skip
	//skipCursor, err := q.All(&items, cursorFields...)
	// Startting from last skipped item, we get data
	//_, err = q.Cursor(skipCursor).Limit(limit).All(&items, cursorFields...)
	// Cover bson items to generic slices items
	genericItems := make([]map[string]interface{}, len(items))
	for index, item := range items {
		doc := item.(bson.M)
		genericItems[index] = createMapFromBsonM(doc)
	}
	return PagedResults{
		Total:           total,
		CurrentPage:     page,
		TotalPage:       pagingInfor.TotalPage,
		PageSize:        len(genericItems),
		NextPage:        pagingInfor.NextPage,
		PreviousPage:    pagingInfor.PreviousPage,
		HasNextPage:     pagingInfor.HasNextPage,
		HasPreviousPage: pagingInfor.HasPreviousPage,
		Items:           genericItems,
	}, nil
}

// GetTotal get all items with paging infor
func (m *mongoHandler) GetTotal(dataname string, filters map[string]interface{}) (int, error) {
	// Make sure connection open
	err := m.GetConnection()
	if err != nil {
		log.Printf("[App.db]: Error during create mongo session: %s\n", err)
		return 0, err
	}
	workingDBSession := m.connection.Copy()
	defer workingDBSession.Close()
	//var cursorFields  []string
	c := workingDBSession.DB(m.database).C(dataname)
	// Get total items by filters
	total, err := c.Find(filters).Count()
	if err != nil {
		log.Printf("[App.db]: Error during couting items: %s\n", err)
		return 0, err
	}

	return total, nil
}

// GetAllItemsNoLimit get all items no limit
func (m *mongoHandler) GetAllItemsNoLimit(dataname string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	err := m.GetConnection()
	if err != nil {
		log.Printf("[App.db]: Error during create mongo session: %s\n", err)
		return nil, err
	}
	workingDBSession := m.connection.Copy()
	defer workingDBSession.Close()
	c := workingDBSession.DB(m.database).C(dataname)
	var items []interface{}
	err = c.Find(filters).All(&items)
	if err != nil {
		log.Printf("[App.db]: Error during get all items: %s\n", err)
		return nil, err
	}
	genericItems := make([]map[string]interface{}, len(items))
	for index, item := range items {
		doc := item.(bson.M)
		genericItems[index] = createMapFromBsonM(doc)
	}
	return genericItems, err
}

func (m *mongoHandler) AddNewItem(dataName string, item map[string]interface{}) (map[string]interface{}, error) {
	// Make sure not modify original map
	willInsertDoc := cloneStringMap(item)
	// Make sure connection open
	err := m.GetConnection()
	if err != nil {
		log.Printf("[App.db]: Error during save %+v\n. %s\n", item, err)
		return willInsertDoc, err
	}
	// Create unique id for item
	if providedID, ok := willInsertDoc["_id"]; !ok || providedID == nil || providedID == "" {
		willInsertDoc["_id"] = bson.NewObjectId()
	}
	if _, ok := willInsertDoc["_id"].(bson.ObjectId); !ok {
		willInsertDoc["_id"], err = createObjectID(willInsertDoc["_id"])
		if err != nil {
			return willInsertDoc, err
		}
	}
	workingDBSession := m.connection.Copy()
	defer workingDBSession.Close()
	c := workingDBSession.DB(m.database).C(dataName)
	err = c.Insert(willInsertDoc)
	if err != nil {
		return item, err
	}
	// return hexid
	returnedID, _ := willInsertDoc["_id"].(bson.ObjectId).MarshalText()
	willInsertDoc["_id"] = string(returnedID)
	return willInsertDoc, err
}

func (m *mongoHandler) RemoveItemByID(dataName string, id interface{}) error {
	// Make sure connection open
	err := m.GetConnection()
	// Make sure to use correct object id
	objectID, err := createObjectID(id)
	if err != nil {
		log.Printf("[App.db]: Error remove item %s. %s\n", id, err)
		return err
	}
	workingDBSession := m.connection.Copy()
	defer workingDBSession.Close()
	c := workingDBSession.DB(m.database).C(dataName)
	response := c.RemoveId(objectID)
	return response
}

func (m *mongoHandler) FindItemByID(dataName string, id interface{}) (map[string]interface{}, error) {
	var data map[string]interface{}
	// Make sure connection open
	err := m.GetConnection()
	if err != nil {
		log.Printf("[App.db]: Error remove item %s. %s\n", id, err)
		return data, err
	}
	// Make sure to use correct object id
	objectID, err := createObjectID(id)
	if err != nil {
		log.Printf("[App.db]: Error during create object id %s. %s\n", id, err)
		return data, err
	}
	workingDBSession := m.connection.Copy()
	defer workingDBSession.Close()
	c := workingDBSession.DB(m.database).C(dataName)
	var found interface{}
	err = c.FindId(objectID).One(&found)
	if err != nil {
		log.Printf("[App.db]: Error find item %s. %s\n", id, err)
		return data, err
	}
	data = createMapFromBsonM(found.(bson.M))
	return data, nil
}

func (m *mongoHandler) FindBy(dataName string, selector map[string]interface{}) (map[string]interface{}, error) {
	var data map[string]interface{}
	// Make sure connection open
	err := m.GetConnection()
	if err != nil {
		return data, err
	}
	workingDBSession := m.connection.Copy()
	defer workingDBSession.Close()
	c := workingDBSession.DB(m.database).C(dataName)
	var found interface{}
	err = c.Find(selector).One(&found)
	if err != nil {
		return data, err
	}
	data = createMapFromBsonM(found.(bson.M))
	return data, nil
}

// Convert bson.M to generic type
func createMapFromBsonM(doc bson.M) map[string]interface{} {
	var message map[string]interface{}
	message = map[string]interface{}(doc)
	// Set id to id string
	if isBsonMContenNonEmptyKey(doc, "_id") {
		objectIDText, _ := doc["_id"].(bson.ObjectId).MarshalText()
		message["_id"] = string(objectIDText)
	}
	return message
}
func isBsonMContenNonEmptyKey(data bson.M, key string) bool {
	val, ok := data[key]
	return ok && val != nil
}

func (m *mongoHandler) UpdateBy(dataName string, selector, update map[string]interface{}) (int, error) {
	// Make sure connection open
	err := m.GetConnection()
	if err != nil {
		log.Printf("[App.db]: Error during get connection for updating item %s. %s\n", selector, err)
		return 0, err
	}
	willUpdateDoc := cloneStringMap(update)
	willSelector := cloneStringMap(selector)
	delete(update, "_id")
	workingDBSession := m.connection.Copy()
	defer workingDBSession.Close()
	c := workingDBSession.DB(m.database).C(dataName)
	rs, err := c.UpdateAll(willSelector, bson.M{"$set": willUpdateDoc})
	return rs.Updated, err
}
func (m *mongoHandler) UpdateByID(dataName string, id interface{}, update map[string]interface{}) error {
	// Make sure connection open
	err := m.GetConnection()
	if err != nil {
		log.Printf("[App.db]: Error during get connection for updating item %s. %s\n", id, err)
		return err
	}
	workingDBSession := m.connection.Copy()
	defer workingDBSession.Close()
	c := workingDBSession.DB(m.database).C(dataName)
	// Make sure to use correct object id
	objectID, err := createObjectID(id)
	if err != nil {
		log.Printf("[App.db]: Error during create object id %s. %s\n", id, err)
		return err
	}
	// Not allow to update id
	willUpdateDoc := cloneStringMap(update)
	delete(willUpdateDoc, "_id")
	response := c.UpdateId(objectID, willUpdateDoc)
	return response
}

// func (m *mongoHandler) UpdateByDeviceAndTokenFirebase(dataName string, userID int, device string, token string) (map[string]interface{}, error) {
// 	// Make sure connection open
// 	err := m.GetConnection()
// 	if err != nil {
// 		log.Printf("[App.db]: Error during get connection for updating item %s.\n", err)
// 		return nil, err
// 	}
// 	workingDBSession := m.connection.Copy()
// 	defer workingDBSession.Close()
// 	c := workingDBSession.DB(m.database).C(dataName)
// 	var found interface{}
// 	err = c.Find(bson.M{device: bson.M{"$in": []string{token}}}).One(&found)
// 	if err != nil {
// 		return nil, err
// 	}
// 	//One
// 	data := createMapFromBsonM(found.(bson.M))
// 	data["userID"] = userID
// 	// Make sure to use correct object id
// 	objectID, err := createObjectID(data["_id"])
// 	if err != nil {
// 		log.Printf("[App.db]: Error during create object id %s. %s\n", data["_id"], err)
// 		return nil, err
// 	}
// 	// Not allow to update id
// 	willUpdateDoc := cloneStringMap(data)
// 	delete(willUpdateDoc, "_id")
// 	response := c.UpdateId(objectID, willUpdateDoc)

// 	var foundRespone interface{}
// 	err = c.FindId(objectID).One(&foundRespone)
// 	if err != nil {
// 		return nil, err
// 	}
// 	dataRes := createMapFromBsonM(foundRespone.(bson.M))
// 	return dataRes, response
// }

// func (m *mongoHandler) DeleteUserTopicDeviceToken(dataName string, info map[string]interface{}) error {
// 	// Make sure connection open
// 	err := m.GetConnection()
// 	if err != nil {
// 		log.Printf("[App.db]: Error during get connection for delete item %s. %s\n", info["userID"], err)
// 		return err
// 	}
// 	workingDBSession := m.connection.Copy()
// 	defer workingDBSession.Close()
// 	c := workingDBSession.DB(m.database).C(dataName)
// 	response := c.Remove(bson.M{"topicUserID": info["userID"], "topic": info["topic"]})
// 	// Make sure to use correct object id
// 	return response
// }

func (m *mongoHandler) RemoveItemBy(dataName string, selector map[string]interface{}) error {
	// Make sure connection open
	err := m.GetConnection()
	if err != nil {
		log.Printf("[App.db]: Error during get connection for RemoveItemBy %s\n", err)
		return err
	}
	workingDBSession := m.connection.Copy()
	defer workingDBSession.Close()
	c := workingDBSession.DB(m.database).C(dataName)
	response := c.Remove(selector)
	// Make sure to use correct object id
	return response
}

func createObjectID(id interface{}) (bson.ObjectId, error) {
	if _, ok := id.(bson.ObjectId); !ok {
		// Incase input is string
		stringID, ok := id.(string)
		if ok {
			if !bson.IsObjectIdHex(stringID) {
				return bson.ObjectId(""), errors.New("Wrong id format")
			}
			return bson.ObjectIdHex(stringID), nil
		}
		bytesID, ok := id.([]byte)
		if !ok {
			return bson.ObjectId(""), errors.New("Unsuported input: only support string and []byte")
		}
		// create a (may be invalid) object type of ObjectId
		var result = bson.ObjectId(bytesID)
		err := result.UnmarshalText(bytesID)
		if err != nil {
			return bson.ObjectId(""), errors.New("Wrong id format")
		}

		return result, nil
	}
	return id.(bson.ObjectId), nil
}

func (m *mongoHandler) createMongoSession() (*mgo.Session, error) {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{m.host + ":" + strconv.Itoa(m.port)},
		Timeout:  60 * time.Second,
		Database: m.autdb,
		Username: m.username,
		Password: m.password,
		MaxIdleTimeMS: m.maxIdleTimeMS,
	}
	// Create a session which maintains a pool of socket connections
	// to our MongoDB.
	mongoSession, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		log.Printf("[App.db]: Error during create mongo session: %s\n", err)
		return nil, err
	}

	return mongoSession, nil
}

// NewMongoHandler create a instance of mongo db
func NewMongoHandler(host, database, authdb, username, password string, port, maxIDLETimeMS int) DatabaseHandler {
	return &mongoHandler{
		host:          host,
		port:          port,
		database:      database,
		autdb:         authdb,
		username:      username,
		password:      password,
		maxIdleTimeMS: maxIDLETimeMS,
	}
}