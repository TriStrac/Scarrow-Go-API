# Database Schema: `scarrow_db` (Refined)

**API Version:** 1.9.0

This document outlines the professional table structures for the Scarrow-Go-API, featuring explicitly named primary keys and a simplified user-device ownership model.

## Table of Contents
1. [device_logs](#device_logs)
2. [devices](#devices)
3. [group_invitations](#group_invitations)
4. [groups](#groups)
5. [message_threads](#message_threads)
6. [messages](#messages)
7. [notifications](#notifications)
8. [otp_codes](#otp_codes)
9. [push_tokens](#push_tokens)
10. [subscription_plans](#subscription_plans)
11. [user_activity_logs](#user_activity_logs)
12. [user_addresses](#user_addresses)
13. [user_profiles](#user_profiles)
14. [user_subscriptions](#user_subscriptions)
15. [users](#users)

---

## Notes on Unused Tables

The following tables exist in the schema but are **not actively used** by the API:

| Table | Purpose | Status | Notes |
|-------|---------|--------|-------|
| `push_tokens` | FCM/APNs push notification tokens | Not used | Long polling (`GET /api/notifications/poll`) is used instead for real-time notifications |
| `device_owners` | Device ownership mapping | Obsolete | Ownership is determined by `user_id` column on `devices` table |

---

### `device_logs`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `log_id` | varchar(36) | NO | NULL | PK |
| `device_id` | varchar(36) | NO | NULL | INDEX |
| `log_type` | varchar(50) | NO | NULL | |
| `pest_type` | varchar(50) | YES | NULL | |
| `frequency_hz` | double | YES | NULL | |
| `duration_seconds` | bigint | YES | NULL | |
| `payload` | text | YES | NULL | |
| `created_at` | datetime(3) | YES | NULL | |

> **Note:** `device_id` stores the ID of the device that detected the pest — if a node detected it, `device_id` = node's ID; if hub detected it, `device_id` = hub's ID.

### `devices`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `device_id` | varchar(36) | NO | NULL | PK |
| `name` | varchar(100) | NO | NULL | |
| `user_id` | varchar(36) | NO | NULL | INDEX (FK to users) |
| `type` | varchar(20) | NO | 'CENTRAL' | |
| `parent_id` | varchar(36) | YES | NULL | INDEX |
| `status` | varchar(50) | YES | 'OFFLINE' | |
| `secret` | varchar(64) | YES | NULL | |
| `lat` | decimal(10,8) | YES | NULL | |
| `lng` | decimal(11,8) | YES | NULL | |
| `node_type` | varchar(50) | YES | NULL | |
| `created_at` | datetime(3) | YES | NULL | |
| `updated_at` | datetime(3) | YES | NULL | |
| `is_deleted` | tinyint(1) | YES | '0' | |
| `deleted_at` | datetime(3) | YES | NULL | INDEX |

### `group_invitations`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `code` | varchar(8) | NO | NULL | PK |
| `group_id` | varchar(36) | NO | NULL | INDEX |
| `created_by` | varchar(36) | NO | NULL | |
| `expires_at` | datetime(3) | NO | NULL | |
| `created_at` | datetime(3) | YES | NULL | |
| `deleted_at` | datetime(3) | YES | NULL | INDEX |

### `groups`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `group_id` | varchar(36) | NO | NULL | PK |
| `name` | varchar(100) | NO | NULL | UNIQUE |
| `owner_id` | varchar(36) | NO | NULL | INDEX |
| `is_deleted` | tinyint(1) | YES | '0' | |
| `created_at` | datetime(3) | YES | NULL | |
| `updated_at` | datetime(3) | YES | NULL | |
| `deleted_at` | datetime(3) | YES | NULL | INDEX |

### `message_threads`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `thread_id` | varchar(36) | NO | NULL | PK |
| `usera_id` | varchar(36) | NO | NULL | INDEX |
| `userb_id` | varchar(36) | NO | NULL | INDEX |
| `created_at` | datetime(3) | YES | NULL | |
| `updated_at` | datetime(3) | YES | NULL | |
| `deleted_at` | datetime(3) | YES | NULL | INDEX |

### `messages`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `message_id` | varchar(36) | NO | NULL | PK |
| `thread_id` | varchar(36) | NO | NULL | INDEX |
| `sender_id` | varchar(36) | NO | NULL | |
| `content` | text | NO | NULL | |
| `is_read` | tinyint(1) | YES | '0' | |
| `created_at` | datetime(3) | YES | NULL | |
| `deleted_at` | datetime(3) | YES | NULL | INDEX |

### `notifications`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `notification_id` | varchar(36) | NO | NULL | PK |
| `user_id` | varchar(36) | NO | NULL | INDEX |
| `title` | varchar(255) | NO | NULL | |
| `message` | text | NO | NULL | |
| `is_read` | tinyint(1) | YES | '0' | |
| `created_at` | datetime(3) | YES | NULL | |
| `updated_at` | datetime(3) | YES | NULL | |
| `deleted_at` | datetime(3) | YES | NULL | INDEX |

### `otp_codes`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `otp_id` | varchar(36) | NO | NULL | PK |
| `identifier` | varchar(100) | NO | NULL | INDEX |
| `destination` | varchar(20) | NO | NULL | |
| `code` | varchar(6) | NO | NULL | |
| `purpose` | varchar(20) | NO | NULL | |
| `payload` | text | YES | NULL | |
| `expires_at` | datetime(3) | NO | NULL | |
| `is_used` | tinyint(1) | YES | '0' | |
| `created_at` | datetime(3) | YES | NULL | |

### `push_tokens`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `token_id` | varchar(36) | NO | NULL | PK |
| `user_id` | varchar(36) | NO | NULL | INDEX |
| `token` | varchar(255) | NO | NULL | UNIQUE |
| `platform` | varchar(50) | YES | NULL | |
| `created_at` | datetime(3) | YES | NULL | |
| `updated_at` | datetime(3) | YES | NULL | |

### `subscription_plans`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `plan_id` | varchar(36) | NO | NULL | PK |
| `name` | varchar(100) | NO | NULL | UNIQUE |
| `description` | text | YES | NULL | |
| `price` | double | YES | NULL | |
| `currency` | varchar(10) | YES | 'PHP' | |
| `duration_days` | bigint | YES | NULL | |
| `created_at` | datetime(3) | YES | NULL | |
| `updated_at` | datetime(3) | YES | NULL | |
| `deleted_at` | datetime(3) | YES | NULL | INDEX |

### `user_activity_logs`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `log_id` | bigint unsigned | NO | NULL | AUTO_INCREMENT (PK) |
| `user_id` | varchar(36) | NO | NULL | INDEX |
| `action` | varchar(255) | NO | NULL | |
| `module` | varchar(100) | YES | NULL | |
| `created_at` | datetime(3) | YES | NULL | |

### `user_addresses`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `address_id` | varchar(36) | NO | NULL | PK |
| `user_id` | varchar(36) | NO | NULL | INDEX |
| `street_name` | varchar(255) | YES | NULL | |
| `baranggay` | varchar(100) | YES | NULL | |
| `town` | varchar(100) | YES | NULL | |
| `province` | varchar(100) | YES | NULL | |
| `zip_code` | varchar(10) | YES | NULL | |

### `user_profiles`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `profile_id` | varchar(36) | NO | NULL | PK |
| `user_id` | varchar(36) | NO | NULL | INDEX |
| `first_name` | varchar(100) | YES | NULL | |
| `middle_name` | varchar(100) | YES | NULL | |
| `last_name` | varchar(100) | YES | NULL | |
| `birth_date` | date | YES | NULL | |
| `phone_number` | varchar(20) | YES | NULL | UNIQUE |

### `user_subscriptions`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `subscription_id` | varchar(36) | NO | NULL | PK |
| `user_id` | varchar(36) | NO | NULL | INDEX |
| `plan_id` | varchar(36) | NO | NULL | |
| `status` | varchar(50) | YES | 'PENDING' | |
| `reference_id` | varchar(255) | YES | NULL | |
| `start_date` | datetime(3) | YES | NULL | |
| `end_date` | datetime(3) | YES | NULL | |
| `created_at` | datetime(3) | YES | NULL | |
| `updated_at` | datetime(3) | YES | NULL | |

### `users`
| Column | Type | Nullable | Default | Extra / Key |
| :--- | :--- | :--- | :--- | :--- |
| `user_id` | varchar(36) | NO | NULL | PK |
| `username` | varchar(100) | NO | NULL | UNIQUE (INDEX) |
| `password` | varchar(255) | NO | NULL | |
| `group_id` | varchar(36) | YES | NULL | INDEX |
| `is_in_group` | tinyint(1) | YES | '0' | |
| `is_head` | tinyint(1) | YES | '0' | |
| `is_verified` | tinyint(1) | YES | '0' | |
| `subscription_status` | varchar(50) | YES | 'PREMIUM' | |
| `is_deleted` | tinyint(1) | YES | '0' | |
| `created_at` | datetime(3) | YES | NULL | |
| `updated_at` | datetime(3) | YES | NULL | |
| `deleted_at` | datetime(3) | YES | NULL | INDEX |
