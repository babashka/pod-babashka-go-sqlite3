#!/usr/bin/env bb

(ns honeysql
  (:require [babashka.deps :as deps]
            [babashka.pods :as pods]))

(deps/add-deps '{:deps {honeysql/honeysql {:mvn/version "1.0.444"}}})

(require '[honeysql.core :as sql])

(pods/load-pod "./pod-babashka-sqlite3")

(require '[pod.babashka.sqlite3 :as sqlite])

(prn (sqlite/execute! "/tmp/foo.db" ["create table if not exists foo (col1 TEXT, col2 TEXT)"]))
(prn (sqlite/execute! "/tmp/foo.db" ["delete from foo"]))
(prn (sqlite/execute! "/tmp/foo.db" ["insert into foo values (?,?)" "foo" "bar"]))

(def sqlmap {:select [:col1 :col2]
             :from   [:foo]
             :where  [:= :col1 "foo"]})

(def sql (sql/format sqlmap))

(prn sql)

(def results (sqlite/query! "/tmp/foo.db" sql))
(prn results)

