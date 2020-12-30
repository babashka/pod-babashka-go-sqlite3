#!/usr/bin/env bb

(ns script
  (:require [babashka.pods :as pods]))

(prn (pods/load-pod "./main"))

(require '[pod.babashka.sqlite3 :as sqlite])

(prn (sqlite/execute! "/tmp/foo.db" ["create table if not exists foo (col1 TEXT, col2 TEXT)"]))
(prn (sqlite/execute! "/tmp/foo.db" ["delete from foo"]))
(prn (sqlite/execute! "/tmp/foo.db" ["insert into foo values (?,?)" "foo" "bar"]))
(prn (sqlite/query!   "/tmp/foo.db" ["select * from foo"]))
