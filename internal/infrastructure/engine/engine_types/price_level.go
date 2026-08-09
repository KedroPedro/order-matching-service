package enginetypes

type OrderStorage interface {
	Add(*EngineOrder)
	Delete()
	GetNodes() func() *EngineOrder
	SetParent(Container)
}

type PriceLevel struct {
	price         int64
	totalQuantity int64
	orders        OrderStorage
	parent        Container
}

func NewPriceLevel(price int64, storage OrderStorage, parent Container) *PriceLevel {
	newPriceLevel := PriceLevel{
		price:         price,
		totalQuantity: 0,
		orders:        storage,
		parent:        parent,
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

func (this *PriceLevel) GetOrders() func() *EngineOrder {
	return this.orders.GetNodes()
}

func (this *PriceLevel) Add(order *EngineOrder) {
	this.orders.Add(order)
}
