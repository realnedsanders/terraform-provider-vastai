package subaccount

// Subaccount sweeper is intentionally omitted.
//
// The Subaccount API does not expose a Delete method. The SubaccountService
// only supports Create and List operations. Without a delete capability,
// there is no way to clean up leaked test resources via sweeper.
//
// Acceptance tests that create subaccounts should document this limitation.
// Subaccounts may need manual cleanup if tests fail.
