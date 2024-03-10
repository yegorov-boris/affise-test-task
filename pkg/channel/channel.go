package channel

func TryRead(c <-chan struct{}) {
	select {
	case <-c:
	default:
	}
}
