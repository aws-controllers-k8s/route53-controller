    // You must use a unique CallerReference string every time you submit a
    // CreateHealthCheck request. CallerReference can be any unique string, for
    // example, a date/timestamp.
    // TODO: Name is not sufficient, since a failed request cannot be retried.
    // We might need to import the `time` package into `sdk.go`
    input.CallerReference = aws.String(getCallerReference())