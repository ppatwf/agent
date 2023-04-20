package leaderapi

// Error response is the response body for any errors that occur
type ErrorResponse struct {
	Error string `json:"error"`
}

// ValueResponse is the response body for endpoints that return a single value.
type ValueResponse struct {
	Value string `json:"value"`
}

// LockCASRequest is the request body for the PATCH /lock/{key} endpoint.
type LockCASRequest struct {
	Old string `json:"old"`
	New string `json:"new"`
}

// LockCASRequest is the response body for the PATCH /lock/{key} endpoint.
type LockCASResponse struct {
	Value   string `json:"value"`
	Swapped bool   `json:"swapped"`
}
