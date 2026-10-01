import type { UsersConfig } from '../stores/config';

/** Comments can remain available to guests without exposing an account entry. */
export function publicAccountsEnabled(users?: UsersConfig): boolean {
  return !!users?.enabled && !!(users.readers?.enabled || users.authors?.enabled);
}
