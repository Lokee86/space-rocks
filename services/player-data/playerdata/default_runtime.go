package playerdata

func NewInMemoryRuntime() (*Runtime, error) {
	accountStore := NewMemoryStore()
	localStore := NewMemoryStore()
	guestStore := NewGuestMemoryStore()

	return NewRuntime(Config{
		Store: NewStoreRouter(accountStore, localStore, guestStore),
	})
}
