package internal

const(
	// controller中restQueue的个数
	LevelSize = 10

	// controller中restQueue队列休息时间步长（10分钟）
	LevelRestStep = uint64(120)

	// 计算频率的比例系数（千分比）
	ScaleFactor = uint64(1000)

	// 默认管理的对象的个数
	DefaultObjCount = uint64(1e6)

	// 被访问的单位时间(单位为秒)，在单位时间内访问的次数即为访问频率
	NodeUnitRestTime = 10
)
