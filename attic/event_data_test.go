package eventdata

//// This is experimental!!! And its not even for MVP anyway!
//
//import (
//	"testing"
//	//"github.com/mitchellh/pointerstructure"
//)
//
////func TestPointerStructure(t *testing.T) {
////
////	data := map[string]interface{}{
////		"alice": 42,
////		"bob": []interface{}{
////			map[string]interface{}{
////				"name": "Bob",
////			},
////		},
////	}
////	value, err := pointerstructure.Get(data, "/bob/0/name")
////	if err != nil {
////		t.Fatal(err)
////	}
////	if diff := deep.Equal(value, "Bob"); diff != nil {
////		t.Fatal(diff)
////	}
////}
//when we go post-MVP, we will start supporting "at least once" delivery, and related concepts
//I would like to design the MVP such that when we do that, we don't have to change the API
//
//func TestPointer(t *testing.T) {
//
//	_, err := Parse("/a/b/c")
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//}
