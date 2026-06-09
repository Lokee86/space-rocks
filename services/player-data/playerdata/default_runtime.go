package playerdata

func NewInMemoryRuntime() (*Runtime, error) {
	accountStore := NewMemoryStore()
	localStore := NewMemoryStore()
	guestStore := NewNoopStore()

	return NewRuntime(Config{
		Store: NewStoreRouter(accountStore, localStore, guestStore),
	})
}
