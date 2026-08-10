package enginetypes

type OrderStorage interface {
	Add(*EngineOrder)
	Delete()
	GetNodes() func() *EngineOrder
	SetParent(Container)
	GetFirstStorageNode() StorageNode
}

type StorageNode interface {
	GetNextStorageNode() StorageNode
	GetStorageNodeValue() *EngineOrder
}

type PriceLevel struct {
	price         int64
	totalQuantity int64
	orders        OrderStorage
	parent        Container
	orderIterator OrderIterator
}

type OrderIterator struct {
	curr StorageNode
}

func NewPriceLevel(price int64, storage OrderStorage, parent Container) *PriceLevel {
	newPriceLevel := PriceLevel{
		price:         price,
		totalQuantity: 0,
		orders:        storage,
		parent:        parent,
		orderIterator: OrderIterator{},
	}

	storage.SetParent(&newPriceLevel)

	return &newPriceLevel
}

func (this PriceLevel) GetLevel() int64 {
	return this.price
}

func (this PriceLevel) GetQuantity() int64 {
	return this.totalQuantity
}

func (this *PriceLevel) Delete() {
	this.parent.Delete()
}

func (this *PriceLevel) GetFirst() *EngineOrder {
	return this.orders.GetNodes()()
}

func (this *PriceLevel) DecreaseQuantity(quantity int64) {
	this.totalQuantity -= quantity
}

func (this *PriceLevel) IncreaseQuantity(quantity int64) {
	this.totalQuantity += quantity
}

func (this *PriceLevel) Add(order *EngineOrder) {
	this.orders.Add(order)
}

func (this *PriceLevel) GetOrdersIterator() *OrderIterator {
	this.orderIterator.curr = this.orders.GetFirstStorageNode()
	return &this.orderIterator
}

func (this *OrderIterator) Next() *EngineOrder {
	order := this.curr.GetStorageNodeValue()
	this.curr = this.curr.GetNextStorageNode()
	return order
}
