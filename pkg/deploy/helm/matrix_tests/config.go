package matrix_tests

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gammazero/deque"
	"github.com/mohae/deepcopy"
	"gopkg.in/yaml.v2"
)

const (
	// Variations add ability to create variants for tests matrix
	ConstantVariation string = "__ConstantVariation__"
	RangeVariation    string = "__RangeVariation___"

	// Item works like variation arguments to include special variants of values
	EmptyItem string = "__EmptyItem__"
)

type FileController struct {
	Prefix string
	TmpDir string
	Queue  *deque.Deque
}

type Node struct {
	Keys []interface{}
	Item interface{}
}

func NewNode(item interface{}) Node {
	return Node{Item: item}
}

func LoadConfiguration(path, prefix, dir string) (FileController, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return FileController{}, fmt.Errorf("formatting path failed: %v", err)
	}

	configurationFile, err := ioutil.ReadFile(absPath)
	if err != nil {
		return FileController{}, fmt.Errorf("read matrix tests configuration file failed: %v", err)
	}

	result := make(map[interface{}]interface{})
	err = yaml.Unmarshal(configurationFile, result)
	if err != nil {
		return FileController{}, fmt.Errorf("configuration unmarshalling error: %v", err)
	}
	filesQueue := deque.Deque{}
	filesQueue.PushBack(result)

	if prefix == "" {
		prefix = "values"
	}

	if dir == "" {
		dir = os.TempDir()
	} else {
		dir, err = filepath.Abs(dir)
		if err != nil {
			return FileController{}, fmt.Errorf("saving values failed: %v", err)
		}
	}
	dir = strings.TrimSuffix(dir, "/")
	_ = os.Mkdir(dir, 0755)

	tmpDir, err := ioutil.TempDir(dir, "")
	if err != nil {
		return FileController{}, fmt.Errorf("tmp directory error: %v", err)

	}
	return FileController{Queue: &filesQueue, Prefix: prefix, TmpDir: tmpDir}, nil
}

func findVariations(nodeData interface{}) ([]interface{}, []interface{}) {
	queue := deque.Deque{}
	queue.PushBack(NewNode(nodeData))

	for queue.Len() > 0 {
		tempNode := queue.PopFront().(Node)

		switch tempNode.Item.(type) {
		case map[interface{}]interface{}:
			mapData := tempNode.Item.(map[interface{}]interface{})
			for key, value := range mapData {
				key := key.(string)

				if key == ConstantVariation || key == RangeVariation {
					return tempNode.Keys, value.([]interface{})
				}
				queue.PushBack(Node{Keys: append(tempNode.Keys, key), Item: value})
			}
		case []interface{}:
			arrayItem := tempNode.Item.([]interface{})
			for index, value := range arrayItem {
				queue.PushBack(Node{Keys: append(tempNode.Keys, index), Item: value})
			}
		}
	}
	return nil, nil
}

func (f *FileController) FindAll() {
	var file interface{}
	counter := 0

	for f.Queue.Len() > counter {
		file = f.Queue.PopFront()

		keys, values := findVariations(file)
		if keys == nil {
			counter += 1
			f.Queue.PushBack(file)
			continue
		}

		for _, item := range values {
			f.Queue.PushBack(formatFile(deepcopy.Copy(file), keys, item, 0))
		}
	}
}

func formatFile(file interface{}, keys []interface{}, resultItem interface{}, counter int) interface{} {
	key := keys[counter]

	switch file.(type) {
	case map[interface{}]interface{}:
		mapFile := file.(map[interface{}]interface{})
		if len(keys)-1 == counter {
			if resultItem == EmptyItem {
				delete(mapFile, key)
			} else {
				mapFile[key] = resultItem
			}
		} else {
			// Recursion call, need to be fixed
			mapFile[key] = formatFile(mapFile[key], keys, resultItem, counter+1)
		}
		return mapFile

	case []interface{}:
		intKey := key.(int)
		arrayFile := file.([]interface{})
		if len(keys)-1 == counter {
			if resultItem == EmptyItem {
				// Delete an element from array
				arrayFile[intKey] = arrayFile[len(arrayFile)-1]
				arrayFile[len(arrayFile)-1] = ""
				arrayFile = arrayFile[:len(arrayFile)-1]
			} else {
				arrayFile[intKey] = resultItem
			}
		} else {
			// Recursion call, need to be fixed
			arrayFile[intKey] = formatFile(arrayFile[intKey], keys, resultItem, counter+1)
		}
		return arrayFile
	}
	return file
}

func (f *FileController) SaveValues() error {
	counter := 1
	for f.Queue.Len() > 0 {
		filename := fmt.Sprintf("%s/%s%v.yaml", f.TmpDir, f.Prefix, counter)
		out, err := yaml.Marshal(f.Queue.PopFront())

		if err != nil {
			return fmt.Errorf("saving values file %s failed: %v", filename, err)
		}

		err = ioutil.WriteFile(filename, out, 0755)
		if err != nil {
			return fmt.Errorf("saving values file %s failed: %v", filename, err)
		}
		counter++
	}
	return nil
}

func (f *FileController) Close() {
	_ = os.RemoveAll(f.TmpDir)
}
