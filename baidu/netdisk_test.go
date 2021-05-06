package baidu

import (
	"encoding/json"
	"testing"
)

func TestPreCreateUploadFile_String(t *testing.T) {
	str := `{
  "return_type": 2,
  "errno": 0,
  "info": {
    "size": 2626327,
    "category": 6,
    "isdir": 0,
    "request_id": 273435691682366413,
    "path": "/baidu/test/test.zip",
    "fs_id": 657059106724647,
    "md5": "60bac7b6464d84fed842955e6126826a",
    "ctime": 1545819399,
    "mtime": 1545819399
  },
  "request_id": 273435691682366413
}`
	var uploadFile PreCreateUploadFile
	err := json.Unmarshal([]byte(str), &uploadFile)
	if err != nil {
		t.Fatalf("unmarshl fail:%+v", err)
	} else {
		t.Logf("unmarshl succ:%s", uploadFile.String())
	}
}

func TestCreateFile_String(t *testing.T) {
	str := `{
    "errno": 0,
    "fs_id": 693789892866840,
    "md5": "7d57c40c9fdb4e4a32d533bee1a4e409",
    "server_filename": "test.txt",
    "category": 4,
    "path": "/apps/appName/test.txt",
    "size": 33818,
    "ctime": 1545969541,
    "mtime": 1545969541,
    "isdir": 0,
    "name": "/apps/appName/test.txt"
}`
	var uploadFile CreateFile
	err := json.Unmarshal([]byte(str), &uploadFile)
	if err != nil {
		t.Fatalf("unmarshl fail:%+v", err)
	} else {
		t.Logf("unmarshl succ:%s", uploadFile.String())
	}
}
