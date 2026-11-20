# Spec Delta

## Purpose

Defines how callers configure the Kafka publisher and consumer: required connection values are explicit, optional delivery and fetch settings are configurable with preserved defaults, and invalid configuration fails before any message is read or written.

## ADDED Requirements

### Requirement: Publisher connection parameters are explicit

Creating a publisher SHALL require broker addresses and a topic as explicit inputs. It SHALL reject an empty broker list or a blank topic with an error before any connection attempt.

#### Scenario: Broker list is empty
- **WHEN** a publisher is created without a broker address
- **THEN** creation fails with an error naming the broker list
- **THEN** no connection is attempted

#### Scenario: Topic is blank
- **WHEN** a publisher is created with a blank topic
- **THEN** creation fails with an error naming the topic
- **THEN** no connection is attempted

#### Scenario: Required values are valid
- **WHEN** a publisher is created with at least one broker and a non-blank topic
- **THEN** creation succeeds and the publisher is ready to write messages

### Requirement: Publisher delivery settings are configurable

A caller SHALL be able to set the required acknowledgement level, maximum delivery attempts, batch size, and batch timeout. Each omitted setting SHALL preserve the behavior that existing callers receive.

#### Scenario: Acknowledgement level is set
- **WHEN** a publisher is created with a required acknowledgement level
- **THEN** writes use that acknowledgement level

#### Scenario: Retry and batching are set
- **WHEN** a publisher is created with a maximum attempts value, a batch size, and a batch timeout
- **THEN** failed writes stop at that attempt limit
- **THEN** writes use those batching values

#### Scenario: Delivery options are omitted
- **WHEN** a publisher is created without delivery options
- **THEN** it keeps the previous acknowledgement, retry, batching, and partition-balancing behavior

### Requirement: Consumer connection parameters are explicit

Creating a consumer SHALL require broker addresses, a topic, and a consumer group as explicit inputs. It SHALL reject an empty broker list, a blank topic, or a blank group with an error before reading any message.

#### Scenario: Broker list is empty
- **WHEN** a consumer is created without a broker address
- **THEN** creation fails with an error naming the broker list
- **THEN** no message is read

#### Scenario: Topic or group is blank
- **WHEN** a consumer is created with a blank topic or a blank group
- **THEN** creation fails with an error naming the blank value
- **THEN** no message is read

#### Scenario: Required values are valid
- **WHEN** a consumer is created with at least one broker, a non-blank topic, and a non-blank group
- **THEN** creation succeeds and the consumer is ready to read messages

### Requirement: Consumer fetch settings are configurable

A caller SHALL be able to set the minimum fetch bytes, maximum fetch bytes, maximum wait, and start offset. Each omitted setting SHALL preserve the behavior that existing callers receive.

#### Scenario: Fetch size is set
- **WHEN** a consumer is created with minimum and maximum fetch byte values
- **THEN** reads use those fetch size bounds

#### Scenario: Wait and start offset are set
- **WHEN** a consumer is created with a maximum wait and a start offset
- **THEN** reads use that wait
- **THEN** a partition without a committed offset starts from that offset

#### Scenario: Fetch options are omitted
- **WHEN** a consumer is created without fetch options
- **THEN** it keeps the previous fetch size, wait, start offset, and commit behavior

### Requirement: Invalid configuration is returned, not raised

Creating a publisher or consumer SHALL reject negative optional values and unsupported start offsets. It SHALL return an error naming the rejected setting, SHALL NOT panic, and SHALL NOT open a connection or create a reader.

#### Scenario: Negative delivery value
- **WHEN** a publisher is created with a negative retry, batch, or timeout value
- **THEN** creation returns an error naming that setting
- **THEN** no connection is attempted

#### Scenario: Negative fetch value
- **WHEN** a consumer is created with a negative fetch size or wait
- **THEN** creation returns an error naming that setting
- **THEN** no reader is created

#### Scenario: Unsupported start offset
- **WHEN** a consumer is created with a start offset other than the supported first, last, or default values
- **THEN** creation returns an error naming the start offset
- **THEN** no reader is created

#### Scenario: Consumer configuration fails
- **WHEN** a consumer is created with invalid configuration
- **THEN** creation returns an error instead of panicking
