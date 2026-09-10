Feature: Capture snippets
  As the owner of a local snippet collection
  I want to add validated code without preparing storage manually
  So that the first reusable snippet is durable as soon as the command succeeds

  Scenario Outline: Add the first snippet at the platform default location
    Given no gosnip data directory or database exists on "<platform>"
    And stdin contains the exact UTF-8 bytes "fmt.Println(\"hello\")\\n"
    When I run "gosnip add Hello --language ' Go ' --tag CLI --tag cli --tag examples"
    Then the command exits successfully
    And stdout is one line containing the new ID, name "Hello", language "go", tags "cli,examples", and update time
    And stderr is empty
    And a current version database exists at the platform default location
    And its directory and files allow access only to the current user where the platform supports permission controls
    And exactly one snippet is stored with name "Hello", language "go", and tags "cli,examples"
    And the stored code is byte-for-byte equal to stdin
    And its creation and update timestamps are the same UTC instant

    Examples:
      | platform |
      | Linux    |
      | macOS    |
      | Windows  |

  Scenario: Accept whitespace-only code without changing it
    Given no gosnip database exists
    And stdin contains valid UTF-8 made only of spaces, tabs, and line endings
    When I add a snippet with a valid name, language, and tags
    Then the command exits successfully
    And the stored code is byte-for-byte equal to stdin

  Scenario: Refuse to wait for code from a terminal
    Given no gosnip data directory or database exists
    And stdin is an interactive terminal
    When I run add with a valid name, language, and tags
    Then the command exits with operational status 1 without waiting for EOF
    And stderr explains how to provide code through a pipe or redirection
    And stdout is empty
    And no data directory, database, backup, or snippet is created

  Scenario Outline: Reject invalid input before creating storage
    Given no gosnip data directory or database exists
    And the add input has "<invalid condition>"
    When I run add
    Then the command exits with operational status 1
    And stderr identifies a validation error and how to correct it
    And stdout is empty
    And no data directory, database, backup, or snippet is created

    Examples:
      | invalid condition                              |
      | zero bytes of code                             |
      | code that is not valid UTF-8                   |
      | code larger than 1 MiB                         |
      | a name empty after trimming                    |
      | a name containing only decimal digits          |
      | a name longer than 100 Unicode characters      |
      | a missing language                             |
      | a language longer than 100 characters          |
      | a language containing a disallowed character   |
      | a tag longer than 100 Unicode characters       |
      | a tag containing a disallowed character        |
      | more than 20 tags                              |

  Scenario: Preserve an existing snippet when a case-folded name conflicts
    Given a current database contains a snippet named "Auth Token"
    And stdin contains valid code
    When I add a snippet named "auth token"
    Then the command exits with operational status 1
    And stderr identifies a name conflict without printing either snippet body
    And stdout is empty
    And the original snippet is unchanged
    And no second case-fold-equivalent name exists

  Scenario: Upgrade a supported older database before adding
    Given a supported older database contains existing snippets
    And no backup for this invocation exists
    And stdin and metadata describe a valid new snippet
    When I run add
    Then a uniquely timestamped backup is created beside the database before its schema changes
    And stderr reports the backup path
    And the schema upgrade and new snippet commit successfully
    And every existing snippet remains unchanged and available
    And the backup remains after success

  Scenario Outline: Preserve an unusable existing database
    Given the configured database has "<database condition>"
    And stdin and metadata describe a valid new snippet
    When I run add
    Then the command exits with operational status 1
    And stderr identifies the database and an actionable recovery step without printing snippet code
    And stdout is empty
    And the original database bytes remain unchanged
    And no snippet is partially created

    Examples:
      | database condition                           |
      | a schema newer than this gosnip supports     |
      | corruption that prevents a safe open         |
      | a migration that fails before commit         |
      | a write failure before commit                |

  Scenario: Refuse an existing database with unsafe permissions
    Given an existing gosnip data directory or database is accessible by other users
    And stdin and metadata describe a valid new snippet
    When I run add
    Then the command exits with operational status 1
    And stderr identifies the unsafe path and explains how to restrict it
    And the command does not change permissions automatically
    And the database and its snippets remain unchanged

  Scenario: Stop waiting for a competing writer
    Given a current database has been continuously locked by another writer
    And stdin and metadata describe a valid new snippet
    When I run add
    Then the command waits no longer than five seconds for the writer
    And the command exits with operational status 1
    And stderr identifies the lock and suggests retrying
    And stdout is empty
    And no partial snippet exists

  Scenario: Resolve concurrent equivalent names without duplication
    Given a current unlocked database has no snippet named "Parser"
    And two add commands concurrently submit valid code as "Parser" and "parser"
    When both commands finish
    Then exactly one command exits successfully with a one-line summary
    And exactly one command exits with operational status 1 and a conflict diagnostic
    And exactly one case-fold-equivalent snippet exists
    And the stored snippet is complete and valid
