package cmap_test

import (
	"cmap/cmap"
	"fmt"
	"strconv"
	"testing"
)

func generateMap() *cmap.ConcurrentMap{
	m := cmap.NewConcurrentMap(32)
	for i := 0 ; i < 10000 ; i ++ {
		str := "test" + strconv.Itoa(i)
		m.Set(str , i)
	}
	return m
}
func TestConcurrentMap(t *testing.T){
	m := generateMap()
	for _ , k := range m.Keys(){
		fmt.Printf("%s\n", k)
	}
	//for k , v := range m.Items(){
	//	fmt.Println(k , v)
	//}
	for i := 0 ; i < 100 ;  i ++ {
		str := "test" + strconv.Itoa(i)
		x, ok := m.Get(str)
		if !ok {
			msg := "can't find number" + str
			panic(msg)
		}
		fmt.Printf("%d ", x.(int))
	}
	fmt.Printf("%d\n", m.Count())

	ok := m.Has("test3333")
	if !ok{
		panic("miss")
	}

}

