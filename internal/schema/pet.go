package schema

type Pet struct {
	PetID  int32 `json:"petId" msgpack:"petId"`
	PetNum int32 `json:"petNum" msgpack:"petNum"`
}
