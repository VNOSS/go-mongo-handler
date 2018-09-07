package db

// PagedResults paged results from db
type PagedResults struct {
	Total           int                      `json:"total"`
	CurrentPage     int                      `json:"currentPage"`
	TotalPage       int                      `json:"totalPage"`
	PageSize        int                      `json:"pageSize"`
	NextPage        int                      `json:"nextPage,omitempty"`
	PreviousPage    int                      `json:"previousPage,omitempty"`
	HasNextPage     bool                     `json:"hasNextPage,omitempty"`
	HasPreviousPage bool                     `json:"hasPreviousPage,omitempty"`
	Items           []map[string]interface{} `json:"items"`
}

// DatabaseHandler defines interface for a database handler
type DatabaseHandler interface {
	GetConnection() error
	CloseConnection()
	GetAllItems(dataname, orderBy, sortBy string, limit, page int, filters map[string]interface{}) (PagedResults, error)
	GetTotal(dataname string, filters map[string]interface{}) (int, error)
	GetAllItemsNoLimit(dataname string, filters map[string]interface{}) ([]map[string]interface{}, error)
	AddNewItem(dataName string, item map[string]interface{}) (map[string]interface{}, error)
	RemoveItemByID(dataName string, id interface{}) error
	RemoveItemBy(dataName string, selector map[string]interface{}) error
	FindItemByID(dataName string, id interface{}) (map[string]interface{}, error)
	FindBy(dataName string, selector map[string]interface{}) (map[string]interface{}, error)
	UpdateBy(dataName string, selector, update map[string]interface{}) (int, error)
	UpdateByID(dataName string, id interface{}, update map[string]interface{}) error
	// UpdateByDeviceAndTokenFirebase(dataName string, userID int, device string, token string) (map[string]interface{}, error)
	// DeleteUserTopicDeviceToken(dataName string, info map[string]interface{}) error
	IsConnecting() bool
}
