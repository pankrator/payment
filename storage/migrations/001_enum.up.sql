BEGIN;

CREATE TYPE transaction_status AS ENUM (
	'approved',
	'reversed',
	'refunded',
    'error');

CREATE TYPE transaction_type AS ENUM (
	'authorize',
	'charge',
	'refund',
    'reversal');

COMMIT;