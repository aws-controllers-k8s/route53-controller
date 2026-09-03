
	// Downgrade transient InvalidChangeBatch failures to recoverable so the
	// controller backs off exponentially instead of going terminal (community#2754).
	err = demoteTransientChangeBatchError(err)
