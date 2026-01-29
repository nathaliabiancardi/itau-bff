package clients

type Bulkhead struct {
	sem chan struct{}
}

func NewBulkhead(limit int) *Bulkhead {
	return &Bulkhead{
		sem: make(chan struct{}, limit),
	}
}

func (b *Bulkhead) Acquire() bool {
	select {
	case b.sem <- struct{}{}:
		return true
	default:
		return false
	}
}

func (b *Bulkhead) Release() {
	<-b.sem
}
