# migrator performance

Back in 2018 I wrote a small benchmark to compare 3 DB migrations frameworks. In 2021 I decided to refresh the results. Also, instead of the deprecated proprietary Ruby framework, I included a new contester (liquibase).

The frameworks and their versions used in 2021 performance benchmark were:

- migrator - version used `v2021.1.0`
- flyway - feature rich Java-based DB migration framework: https://flywaydb.org - version used `Flyway Teams Edition 7.14.0 by Redgate`
- liquibase - feature rich Java-based DB migration framework: https://liquibase.org - version used `Liquibase Community 4.4.3 by Datical`

# 2021 results

Compared to 2018 the tests are single tenant with 10k migrations. This is because somewhere between 2018 and 2021 [multiple schemas](https://flywaydb.org/documentation/learnmore/faq.html#multiple-schemas) stopped working in flyway. Also, liquibase doesn't support natively multiple schemas.

You can play around with performance benchmarks yourself. See `test/performance/test.sh` for migrator, `test/performance/flywaydb-test.sh` for flyway, and `test/performance/liquibase-test.sh` for liquibase.

_Note: These are simple tests/helpers scripts and not full blown DB migration benchmark tools. I didn't spend too much time making them look slick. They do their job quite well though._

Results are an average of a few runs. All benchmarks were run on my MacBook.

Applying first 10k migrations on an empty database:

| rank | framework | number of migrations (before - after) | time (s) | memory (MB) |
| ---- | --------- | ------------------------------------- | -------- | ----------- |
| 1.   | migrator  | 0 - 10000                             | 31       | 47          |
| 2.   | liquibase | 0 - 10000                             | 298      | 385         |
| 3.   | flyway    | 0 - 10000                             | 540      | 265         |

And then, on top of existing 10k migrations, applying 10k new ones:

| rank | framework | number of migrations (before - after) | time (s) | memory (MB) |
| ---- | --------- | ------------------------------------- | -------- | ----------- |
| 1.   | migrator  | 10000 - 20000                         | 43       | 79          |
| 2.   | flyway    | 10000 - 20000                         | 816      | 324         |
| 3.   | liquibase | 10000 - 20000                         | 1533     | 660         |

migrator is orders of magnitude better than flyway and liquibase.

flyway was slower in the first test, however behaves better than liquibase in the second test.

Because multiple schemas stopped working in flyway and liquibase doesn't support them natively there wasn't too much sense in doing more comparison benchmarks.

Instead, I ran additional multi-tenant/multiple-schema migrator benchmarks.

# migrator's performance showcase

You can use `test/performance/test.sh` to run any simulation you want. You can also use it to simulate adding new migrations (so called append mode) - scroll to the bottom of that script to see a comment showing you how to do this.

I prepared 2 multi-tenant simulations:

1. 1000 tenants, 20 SQL files in each version, 1000 \* 20 = 20k migrations to apply in each version
2. 500 tenants, 100 SQL files in each version, 500 \* 100 = 50k migrations to apply in each version

## 1000 tenants

Execution time is growing slightly with every new version. The memory consumption grows proportionally to how many migrations are in the database. This is because migrator fetches all migrations from database to compute which migrations were already applied and which are to be applied.

| version | number of migrations (before - after) | time (s) | memory (MB) |
| ------- | ------------------------------------- | -------- | ----------- |
| 1       | 0 - 21001                             | 57       | 66          |
| 2       | 21001 - 41001                         | 58       | 86          |
| 3       | 41001 - 61001                         | 56       | 101         |
| 4       | 61001 - 81001                         | 62       | 165         |
| 5       | 81001 - 101001                        | 62       | 175         |
| 6       | 101001 - 121001                       | 59       | 242         |
| 7       | 121001 - 141001                       | 71       | 280         |
| 8       | 141001 - 161001                       | 68       | 300         |
| 9       | 161001 - 181001                       | 70       | 324         |
| 10      | 181001 - 201001                       | 69       | 380         |

## 500 tenants

Similarly to 1000 tenants, in 500 tenants simulation execution time is growing slightly with every new version. The memory consumption grows proportionally to how many migrations are in the database.

| version | number of migrations (before - after) | time (s) | memory (MB) |
| ------- | ------------------------------------- | -------- | ----------- |
| 1       | 0 - 50501                             | 167      | 126         |
| 2       | 50501 - 100501                        | 170      | 140         |
| 3       | 100501 - 150501                       | 167      | 218         |
| 4       | 150501 - 200501                       | 181      | 292         |
| 5       | 200501 - 250501                       | 178      | 396         |

## Summary

Based on both simulations we can see that migrator under any load is stable and behaves very predictable:

- applied ~300 migrations a second (creating tables, inserting multiple rows, database running on the same machine = my MacBook)
- consumed ~2.5MB memory for every 1k migrations

# 2018

Keeping 2018 results for historical purposes.

_Note: For all 3 DB frameworks I used multiple schemas benchmark. Unfortunately, back in 2018, I didn't commit flyway tests and its impossible now to recreate how multiple schemas were actually set up._

Execution times were the following:

| # tenants | # existing migrations | # migrations to apply | migrator | ruby | flyway |
| --------- | --------------------- | --------------------- | -------- | ---- | ------ |
| 10        | 0                     | 10001                 | 154s     | 670s | 2360s  |
| 10        | 10001                 | 20                    | 2s       | 455s | 340s   |

migrator was the undisputed winner.

The Ruby framework had the undesired functionality of making a DB call each time to check if given migration was already applied. migrator fetched all applied migrations at once and compared them in memory. That was the primary reason why migrator was so much better in the second test.

flyway results were... very surprising. I was so shocked that I had to re-run flyway as well as all other tests. Yes, flyway was 15 times slower than migrator in the first test. In the second test flyway was faster than Ruby. Still a couple orders of magnitude slower than migrator.

The other thing to consider is the fact that migrator is written in go which is known to be much faster than Ruby and Java.
