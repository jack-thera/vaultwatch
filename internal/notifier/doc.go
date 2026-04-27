// Package notifier provides types and utilities for building and dispatching
// secret-expiration notifications in vaultwatch.
//
// The core workflow:
//
//  1. The monitor collects []SecretStatus values (path, TTL, expiry time).
//  2. A Dispatcher evaluates each status against a configurable TTL threshold.
//  3. Statuses below the threshold are converted to Notification values with
//     an appropriate Level (INFO / WARNING / CRITICAL).
//  4. The Dispatcher forwards the notifications to every registered Sender.
//
// Senders implement the Sender interface and can target stdout, a webhook,
// email, PagerDuty, etc.  The built-in StdoutSender writes formatted lines
// to any io.Writer.
//
// Example:
//
//	sender := notifier.NewStdoutSender(os.Stdout)
//	d := notifier.NewDispatcher(24*time.Hour, logger, sender)
//	_ = d.Dispatch(statuses)
package notifier
