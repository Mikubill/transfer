package methods

import (
	"io"
	"sync"
)

func monitor(w *io.PipeWriter, sig *sync.WaitGroup) {
	sig.Wait()
	_ = w.Close()
}
