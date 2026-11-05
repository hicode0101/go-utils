package main

import (
	"fmt"

	"github.com/hicode0101/go-utils/utils"
)

// idgenDemo ID 生成与随机值：Id / Random / 需要构造的 Snowflake、Hashids、WeightRandom。
func idgenDemo() {
	banner("Id ID 生成")
	fmt.Println("SimpleUUID() =", utils.Id.SimpleUUID(), "（32 位无连字符）")
	fmt.Println("ObjectId() =", utils.Id.ObjectId(), "（Mongo 风格）")
	fmt.Println("NanoId(12) =", utils.Id.NanoId(12))
	fmt.Println("SnowflakeNextId() =", utils.Id.SnowflakeNextId(), "（包级默认单例）")
	fmt.Println("SnowflakeNextIdStr() =", utils.Id.SnowflakeNextIdStr())

	// 多节点部署时用 NewSnowflake 指定机器号/数据中心号（0~31），避免重复
	sf := utils.NewSnowflake(1, 1)
	fmt.Println("NewSnowflake(1, 1).NextId() =", sf.NextId())

	// Hashids：把一组数字编成短 ID（防遍历），可无损解码
	hid, _ := utils.NewHashids("my-salt", 8).Encode(1, 2, 3)
	nums, _ := utils.NewHashids("my-salt", 8).Decode(hid)
	fmt.Printf("NewHashids(\"my-salt\", 8).Encode(1, 2, 3) = %s -> Decode = %v\n", hid, nums)

	banner("Random 随机值")
	fmt.Println("RandomInt(1, 100) =", utils.Random.RandomInt(1, 100))
	fmt.Println("RandomInt64(1, 1000000) =", utils.Random.RandomInt64(1, 1000000))
	fmt.Println("RandomFloat(0, 1) =", utils.Random.RandomFloat(0, 1))
	fmt.Println("RandomBool() =", utils.Random.RandomBool())
	fmt.Println("RandomString(8) =", utils.Random.RandomString(8))
	fmt.Println("RandomNumbers(6) =", utils.Random.RandomNumbers(6), "（纯数字验证码）")
	fmt.Println(`RandomStringWithoutStr(8, "0O1l") =`, utils.Random.RandomStringWithoutStr(8, "0O1l"))
	fmt.Println(`RandomChar("ABCDEF") =`, string(utils.Random.RandomChar("ABCDEF")))
	fmt.Println("RandomChinese() =", utils.Random.RandomChinese(), "| RandomBytes(6) =", fmt.Sprintf("%x", utils.Random.RandomBytes(6)))

	// 加权随机：权重 1:2:3:4 对应 A/B/C/D
	wr := utils.NewWeightRandom([]float64{1, 2, 3, 4}, []string{"A", "B", "C", "D"})
	fmt.Println("NewWeightRandom(...).Next() =", wr.Next(), wr.Next(), wr.Next())
}
