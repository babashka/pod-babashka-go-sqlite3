#!/usr/bin/env bb

(ns script
  (:require [babashka.pods :as pods]
            [clojure.java.io :as io]
            [clojure.test :as t :refer [deftest is testing]]))

(prn (pods/load-pod "./pod-babashka-go-sqlite3"))

(require '[pod.babashka.go-sqlite3 :as sqlite])

(.delete (io/file "/tmp/foo.db"))

(prn (sqlite/execute! "/tmp/foo.db" ["create table if not exists foo (the_text TEXT, the_int INTEGER, the_real REAL, the_blob BLOB, the_json JSON)"]))
(prn (sqlite/execute! "/tmp/foo.db" ["delete from foo"]))

(def png (java.nio.file.Files/readAllBytes (.toPath (io/file "resources/babashka.png"))))

(prn (sqlite/execute! "/tmp/foo.db" ["insert into foo (the_text, the_int, the_real, the_blob, the_json) values (?,?,?,?,?)" "foo" 1 3.14 png "{\"bar\": \"hello\"}"]))
(prn (sqlite/execute! "/tmp/foo.db" ["insert into foo (the_text, the_int, the_real) values (?,?,?)" "foo" 2 1.5]))

(testing "multiple results"
  (prn (sqlite/execute! "/tmp/foo.db"
                        ["insert into foo (the_text, the_int, the_real) values (?,?,?), (?,?,?)"
                         "bar" 3 1.5
                         "baz" 4 1.5])))

(def results (sqlite/query "/tmp/foo.db" ["select * from foo order by the_int asc"]))

(def results-min-png (mapv #(dissoc % :the_blob :the_json) results))

(def expected [{:the_int 1, :the_real 3.14, :the_text "foo"}
               {:the_int 2, :the_real 1.5, :the_text "foo"}
               {:the_int 3, :the_real 1.5, :the_text "bar"}
               {:the_int 4, :the_real 1.5, :the_text "baz"}])

(deftest results-test
  (is (= expected results-min-png)))

(def direct-results (sqlite/query "/tmp/foo.db" "select * from foo order by the_int asc"))

(def direct-results-min-png (mapv #(dissoc % :the_blob :the_json) direct-results))

(deftest direct-results-test
  (is (= expected direct-results-min-png)))

(deftest bytes-roundtrip
  (is (= (count png) (count (get-in results [0 :the_blob])))))

(def json-field-result (sqlite/query "/tmp/foo.db" ["select the_json->>'$.bar' as bar from foo where the_json is not null"]))

(deftest json-field-test
  (is (= [{:bar "hello"}] json-field-result)))

(deftest error-test
  (is (thrown-with-msg?
       Exception #"no such column: non_existing"
       (sqlite/query "/tmp/foo.db" ["select non_existing from foo"])))
  (is (thrown-with-msg?
       Exception #"unexpected query type, expected a string or a vector"
       (sqlite/query "/tmp/foo.db" 42)))
  (is (thrown-with-msg?
       Exception #"the sqlite connection must be a string"
       (sqlite/query nil "select * from foo order by the_int asc"))))

(deftest fts50-test
  (sqlite/execute! "/tmp/foo.db" ["CREATE VIRTUAL TABLE email USING fts5(sender, title, body)"])
  (sqlite/execute! "/tmp/foo.db" ["INSERT INTO email VALUES('foo', 'bar', 'baz')"])
  (is (= [{:sender "foo" :title "bar" :body "baz"}]
         (sqlite/query "/tmp/foo.db"
                       ["SELECT * FROM email WHERE email MATCH 'baz';"]))))

(let [{:keys [:fail :error]} (t/run-tests)]
  (System/exit (+ fail error)))
