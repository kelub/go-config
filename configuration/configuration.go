package configuration

type Configer interface {
	//设置键值
	Put(key string, value []byte) error
	//得到键对应值 并返回最后更新 index
	Get(key string) ([]byte, uint64, error)
	//得到键前缀为 prefix 的所有键值对 并返回最后更新的键的 index
	List(prefix string) (map[string][]byte, uint64, error)
	//删除键为 key 的值
	Delete(key string) error
}

//func RunServer(){
//	wg := sync.WaitGroup{}
//	wg.Add(1)
//
//}
//
//func Init(){
//
//}
