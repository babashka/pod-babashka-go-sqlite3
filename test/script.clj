#!/usr/bin/env bb

(ns script
  (:require [babashka.pods :as pods]
            [clojure.test :as t :refer [deftest is]]))

(prn (pods/load-pod "./pod-babashka-sqlite3"))

(require '[pod.babashka.sqlite3 :as sqlite])

(prn (sqlite/execute! "/tmp/foo.db" ["create table if not exists foo (col1 TEXT, col2 TEXT)"]))
(prn (sqlite/execute! "/tmp/foo.db" ["delete from foo"]))
(prn (sqlite/execute! "/tmp/foo.db" ["insert into foo values (?,?)" "foo" "bar"]))
(def results (sqlite/query!   "/tmp/foo.db" ["select * from foo"]))
(prn results)

(deftest results-test
  (is (= [{:col1 "foo", :col2 "bar"}] results)))

(let [{:keys [:fail :error]} (t/run-tests)]
  (System/exit (+ fail error)))
