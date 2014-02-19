package mongoproxy

import (
	"fmt"
	"testing"
	"time"

	"github.com/ParsePlatform/go.freeport"
	"github.com/ParsePlatform/go.mgotest"

	. "launchpad.net/gocheck"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type proxyTest struct {
	MongoServer *mgotest.Server
	Proxy       *Proxy
}

var _ = Suite(&proxyTest{})

func (p *proxyTest) SetUpSuite(c *C) {
	c.Log("starting mongoserver")
	p.MongoServer = mgotest.NewStartedServer(c)
	c.Log("started mongoserver")
}

func (p *proxyTest) SetUpTest(c *C) {
	port, err := freeport.Get()
	if err != nil {
		c.Fatal(err)
	}
	p.Proxy = &Proxy{
		ListenAddr:          portToURL(port),
		MongoAddr:           portToURL(p.MongoServer.Port),
		DispatchQueueLen:    1000,
		MaxMongoConnections: 5,
		DispatcherTimeout:   5 * time.Second,
		ListenerTimeout:     5 * time.Second,
	}
	c.Log("starting proxy")
	go p.Proxy.Start()
	c.Log("proxy started")
}

func (p *proxyTest) TearDownTest(c *C) {
	p.Proxy.Stop()
}

func portToURL(port int) string {
	return fmt.Sprintf("127.0.0.1:%d", port)
}

func (p *proxyTest) CreateSession(c *C) *mgo.Session {
	session, err := mgo.Dial(p.Proxy.ListenAddr)
	if err != nil {
		c.Fatal(err)
	}
	session.SetSafe(&mgo.Safe{FSync: true})
	return session
}

func (p *proxyTest) TestSimpleCRUD(c *C) {
	session := p.CreateSession(c)
	defer session.Close()
	collection := session.DB("test").C("coll1")
	data := map[string]interface{}{
		"_id":  1,
		"name": "abc",
	}
	err := collection.Insert(data)
	c.Check(err, IsNil)
	n, err := collection.Count()
	c.Check(err, IsNil)
	c.Check(n, Equals, 1)
	{
		result := make(map[string]interface{})
		collection.Find(bson.M{"_id": 1}).One(&result)
		c.Check(result["name"], Equals, "abc")
	}
	{
		result := make(map[string]interface{})
		collection.Find(bson.M{"_id": 2}).One(&result)
		c.Check(len(result), Equals, 0)
	}
	err = collection.DropCollection()
	c.Check(err, IsNil)
}

func (p *proxyTest) TestIndexCreation(c *C) {
	index := mgo.Index{
		Key:        []string{"lastname", "firstname"},
		Unique:     true,
		DropDups:   true,
		Background: true, // See notes.
		Sparse:     true,
	}
	session := p.CreateSession(c)
	collection := session.DB("test").C("coll2")
	err := collection.EnsureIndex(index)
	c.Check(err, IsNil)
	session.Close()
	// Recreate session as mgo caches indexes. This will force it to lookup mongo
	session = p.CreateSession(c)
	collection = session.DB("test").C("coll2")
	indexes, err := collection.Indexes()
	c.Check(err, IsNil)
	c.Check(indexes, HasLen, 2) // mongo creates "_id" index by default
	session.Close()
}

func (p *proxyTest) TestCreateRemove(c *C) {
	session := p.CreateSession(c)
	collection := session.DB("test").C("coll3")
	// defer's are subtle.
	// the lower defer is called first :)
	defer session.Close()
	defer collection.DropCollection()
	err := collection.Insert(createData(1, 10)...)
	c.Check(err, IsNil)
	count, err := collection.Count()
	c.Check(err, IsNil)
	c.Check(count, Equals, 10)
	err = collection.Remove(map[string]interface{}{"_id": 2})
	c.Check(err, IsNil)
	count, err = collection.Count()
	c.Check(err, IsNil)
	c.Check(count, Equals, 9)
	err = collection.Remove(map[string]interface{}{"_id": 3})
	c.Check(err, IsNil)
	count, err = collection.Count()
	c.Check(err, IsNil)
	c.Check(count, Equals, 8)
}

func (p *proxyTest) TestUpdates(c *C) {
	session := p.CreateSession(c)
	collection := session.DB("test").C("coll3")
	// defer's are subtle.
	// the lower defer is called first :)
	defer session.Close()
	defer collection.DropCollection()
	data := map[string]interface{}{
		"_id":  1,
		"city": "oslo",
	}
	err := collection.Insert(data)
	c.Check(err, IsNil)

	err = collection.Update(map[string]interface{}{"_id": 1},
		map[string]interface{}{"name": "kona", "city": "sanfrancisco"})
	c.Check(err, IsNil)
	result := make(map[string]interface{})
	err = collection.Find(map[string]interface{}{"_id": 1}).One(&result)
	c.Check(err, IsNil)
	c.Check(result["name"], DeepEquals, "kona")
	c.Check(result["city"], DeepEquals, "sanfrancisco")

}

func createData(startIndex int, n int) []interface{} {
	result := make([]interface{}, n)
	for i := 0; i < n; i++ {
		result[i] = map[string]interface{}{
			"_id":  i,
			"time": time.Now(),
		}
	}
	return result
}

func (p *proxyTest) TearDownSuite(c *C) {
	p.MongoServer.Stop()
}

func Test(t *testing.T) { TestingT(t) }
