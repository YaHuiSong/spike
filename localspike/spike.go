package localspike

// LocalSpike stores the limit and sales
type LocalSpike struct {
	Stock, Sales int64
}

//DeductStock : 本地扣库存,返回bool值
func (spike *LocalSpike) DeductStock() bool {
	spike.Sales = spike.Sales + 1
	return spike.Sales <= spike.Stock
}