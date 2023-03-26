package wordpress

import (
	"fmt"
	"net/http"
)

type AvatarURLS struct {
	Size24 string `json:"24,omitempty"`
	Size48 string `json:"48,omitempty"`
	Size96 string `json:"96,omitempty"`
}
type User struct {
	collection *UsersCollection `json:"-"`

	ID                int                    `json:"id,omitempty"`
	AvatarURL         string                 `json:"avatar_url,omitempty"`
	AvatarURLs        AvatarURLS             `json:"avatar_urls,omitempty"`
	Capabilities      map[string]interface{} `json:"capabilities,omitempty"`
	Description       string                 `json:"description,omitempty"`
	Email             string                 `json:"email,omitempty"`
	ExtraCapabilities map[string]interface{} `json:"extra_capabilities,omitempty"`
	FirstName         string                 `json:"first_name,omitempty"`
	LastName          string                 `json:"last_name,omitempty"`
	Link              string                 `json:"link,omitempty"`
	Name              string                 `json:"name,omitempty"`
	Nickname          string                 `json:"nickname,omitempty"`
	RegisteredDate    string                 `json:"registered_date,omitempty"`
	Roles             []string               `json:"roles,omitempty"`
	Slug              string                 `json:"slug,omitempty"`
	URL               string                 `json:"url,omitempty"`
	Username          string                 `json:"username,omitempty"`
	Password          string                 `json:"password,omitempty"`
}

type UsersCollection struct {
	client *Client
	url    string
}

func (entity *User) setCollection(col *UsersCollection) {
	entity.collection = col
}
func (entity *User) Meta() *MetaCollection {
	if entity.collection == nil {
		// missing user.collection parent. Probably User struct was initialized manually.
		_warning("Missing parent post collection")
		return nil
	}
	return &MetaCollection{
		client:     entity.collection.client,
		parent:     entity,
		parentType: CollectionUsers,
		url:        fmt.Sprintf("%v/%v/%v", entity.collection.url, entity.ID, CollectionMeta),
	}
}
func (col *UsersCollection) Me(params interface{}) (*User, *http.Response, []byte, error) {
	url := fmt.Sprintf("%v/me", col.url)
	var user User
	resp, body, err := col.client.Get(url, params, &user)
	return &user, resp, body, err
}
func (col *UsersCollection) List(params interface{}) ([]User, *http.Response, []byte, error) {
	var users []User
	resp, body, err := col.client.List(col.url, params, &users)
	// set collection object for each entity which has sub-collection
	for i := range users {
		users[i].setCollection(col)
	}

	return users, resp, body, err
}
func (col *UsersCollection) Create(new *User) (*User, *http.Response, []byte, error) {
	var created User
	resp, body, err := col.client.Create(col.url, new, &created)

	created.setCollection(col)

	return &created, resp, body, err
}
func (col *UsersCollection) Get(id int, params interface{}) (*User, *http.Response, []byte, error) {
	var entity User
	entityURL := fmt.Sprintf("%v/%v", col.url, id)
	resp, body, err := col.client.Get(entityURL, params, &entity)

	// set collection object for each entity which has sub-collection
	entity.setCollection(col)

	return &entity, resp, body, err
}
func (col *UsersCollection) Update(id int, post *User) (*User, *http.Response, []byte, error) {
	var updated User
	entityURL := fmt.Sprintf("%v/%v", col.url, id)
	resp, body, err := col.client.Update(entityURL, post, &updated)

	// set collection object for each entity which has sub-collection
	updated.setCollection(col)

	return &updated, resp, body, err
}
func (col *UsersCollection) Delete(id int, params interface{}) (*User, *http.Response, []byte, error) {
	var deleted User
	entityURL := fmt.Sprintf("%v/%v", col.url, id)
	resp, body, err := col.client.Delete(entityURL, params, &deleted)

	// set collection object for each entity which has sub-collection
	deleted.setCollection(col)

	return &deleted, resp, body, err
}
