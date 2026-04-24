import pandas as pd
import numpy as np
import random
from pyspark.sql import SparkSession
from pyspark.sql.functions import *
from pyspark.sql.types import *
from pyspark.sql.window import Window
from pyspark.ml.feature import (
    Tokenizer, StopWordsRemover, CountVectorizer, IDF, StringIndexer, 
    NGram, RegexTokenizer, HashingTF, Word2Vec
)
from pyspark.ml.classification import LinearSVC, LogisticRegression, RandomForestClassifier
from pyspark.ml import Pipeline
from pyspark.ml.evaluation import MulticlassClassificationEvaluator, BinaryClassificationEvaluator
from pyspark.ml.tuning import CrossValidator, ParamGridBuilder
from pyspark.mllib.evaluation import MulticlassMetrics

spark = SparkSession.builder.appName("AdvancedLogAnalysis").config("spark.sql.adaptive.enabled", "true").getOrCreate()

INFO_PATTERNS = [
    "user login successful for id={id}",
    "service {svc} started on port {port}",
    "backup completed successfully in {time}s",
    "cache cleared for session {sess}",
    "connection established to db_primary",
    "request processed in {ms}ms with status 200",
    "system startup completed",
    "file uploaded successfully",
    "heartbeat received from node-{node}"
]

WARN_PATTERNS = [
    "disk usage is above {pct}% threshold",
    "memory pressure detected, swapping enabled",
    "deprecated api call used in module {mod}",
    "retry attempt {n} for connection timeout",
    "slow query detected: execution time {time}ms",
    "certificate expires in {days} days",
    "high cpu load detected: {load}%",
    "connection pool exhausted, waiting for free slot"
]

ERROR_PATTERNS = [
    "java.lang.NullPointerException at {cls}.{method}",
    "fatal error: database connection refused",
    "segmentation fault in module {mod}",
    "out of memory exception: heap space",
    "permission denied: cannot write to /var/log/{file}",
    "critical failure in service {svc}: exit code 1",
    "timeout exceeded while waiting for response",
    "uncaught exception: thread crashed"
]

data = []
for _ in range(2000):
    data.append((random.choice(INFO_PATTERNS).format(id=random.randint(100,999), svc="httpd", port=80, time=random.randint(10,100), sess="xyz", ms=random.randint(50,500), node=random.randint(1,10)), "INFO"))
    data.append((random.choice(WARN_PATTERNS).format(pct=random.randint(80,95), mod="legacy_v1", n=random.randint(1,5), time=random.randint(1000,5000), days=random.randint(1,30), load=random.randint(90,100)), "WARN"))
    data.append((random.choice(ERROR_PATTERNS).format(cls="UserService", method="process", mod="net_driver", file="sys.log", svc="db_connector"), "ERROR"))

df = pd.DataFrame(data, columns=['text', 'label'])
df_spark = spark.createDataFrame(df)
df_spark.write.mode("overwrite").parquet("logs_dataset.parquet")
df_spark = spark.read.parquet("logs_dataset.parquet")

print("=== DATASET EXPLORATION ===")
df_spark.groupBy("label").count().orderBy("count", ascending=False).show()
df_spark.select("text").describe().show()

df_spark.withColumn("text_len", length(col("text"))).agg(
    min("text_len"), max("text_len"), avg("text_len"), stddev("text_len")
).toPandas().T

print("=== N-GRAM ANALYSIS ===")
regex_tokenizer = RegexTokenizer(inputCol="text", outputCol="words", pattern="\\w+", minTokenLength=2)
ngram = NGram(n=2, inputCol="words", outputCol="bigrams")
words_ngrams = ngram.transform(regex_tokenizer.transform(df_spark))
words_ngrams.select(explode("bigrams").alias("ngram"), "label").groupBy("ngram").count().orderBy(desc("count")).limit(20).show()

print("=== FEATURE ENGINEERING ===")
tokenizer = Tokenizer(inputCol="text", outputCol="words")
remover = StopWordsRemover(inputCol="words", outputCol="filtered_words")
count_vec = CountVectorizer(inputCol="filtered_words", outputCol="raw_features", vocabSize=10000, minDF=5.0)
idf = IDF(inputCol="raw_features", outputCol="features")
label_indexer = StringIndexer(inputCol="label", outputCol="label_index")

print("=== MODEL TRAINING ===")
svc = LinearSVC(labelCol="label_index", featuresCol="features", maxIter=100, regParam=0.1)
lr = LogisticRegression(labelCol="label_index", featuresCol="features", maxIter=100, regParam=0.01)
rf = RandomForestClassifier(labelCol="label_index", featuresCol="features", numTrees=50)

pipeline_svc = Pipeline(stages=[label_indexer, tokenizer, remover, count_vec, idf, svc])
pipeline_lr = Pipeline(stages=[label_indexer, tokenizer, remover, count_vec, idf, lr])

train_df, test_df = df_spark.randomSplit([0.8, 0.2], seed=42)
train_val, val_df = train_df.randomSplit([0.875, 0.125], seed=42)

model_svc = pipeline_svc.fit(train_df)
model_lr = pipeline_lr.fit(train_df)

preds_svc = model_svc.transform(test_df)
preds_lr = model_lr.transform(test_df)

evaluator_acc = MulticlassClassificationEvaluator(labelCol="label_index", predictionCol="prediction", metricName="accuracy")
evaluator_f1 = MulticlassClassificationEvaluator(labelCol="label_index", predictionCol="prediction", metricName="f1")
evaluator_wf1 = MulticlassClassificationEvaluator(labelCol="label_index", predictionCol="prediction", metricName="weightedPrecision")

print("SVC Results:", evaluator_acc.evaluate(preds_svc), evaluator_f1.evaluate(preds_svc), evaluator_wf1.evaluate(preds_svc))
print("LR Results:", evaluator_acc.evaluate(preds_lr), evaluator_f1.evaluate(preds_lr), evaluator_wf1.evaluate(preds_lr))

print("=== CONFUSION MATRIX ===")
preds_svc.groupBy("label", "prediction").count().orderBy("label", "prediction").show()

print("=== HYPERPARAMETER TUNING ===")
paramGrid = ParamGridBuilder() \
    .addGrid(count_vec.vocabSize, [5000, 10000, 15000]) \
    .addGrid(count_vec.minDF, [2.0, 5.0, 10.0]) \
    .addGrid(svc.regParam, [0.01, 0.1, 1.0]) \
    .addGrid(svc.maxIter, [100, 200]) \
    .build()

crossval = CrossValidator(estimator=pipeline_svc,
                         estimatorParamMaps=paramGrid,
                         evaluator=evaluator_f1,
                         numFolds=3)
cv_model = crossval.fit(train_val)
cv_preds = cv_model.transform(val_df)
print("CV Best F1:", evaluator_f1.evaluate(cv_preds))

print("=== FEATURE IMPORTANCE (Top 20) ==="
cv_model.stages[-1].coefficients.toArray()
feature_names = cv_model.stages[3].vocabulary
importance_df = spark.createDataFrame([
    (i, float(coeff), feature_names[i]) for i, coeff in enumerate(cv_model.stages[-1].coefficients)
], ["feature_id", "importance", "feature_name"])
importance_df.orderBy(desc("importance")).limit(20).show(truncate=False)

print("=== PRODUCTION PIPELINE ===")
final_pipeline = cv_model.bestModel
final_pipeline.save("log_classifier_pipeline")

print("=== REAL-TIME PREDICTION ===")
test_logs = [
    "java.lang.NullPointerException at com.example.service.UserService.process(UserService.java:45)",
    "database connection timeout after 3000ms",
    "segmentation fault (core dumped) in module network_driver",
    "disk usage is above 85% threshold on partition /dev/sda1",
    "memory pressure detected: swapping enabled, performance may degrade",
    "deprecated api call used in legacy_module_v1, please update to v2",
    "user login successful for admin_id=1023 from ip 192.168.1.5",
    "scheduled backup completed successfully in 12m 30s",
    "service httpd started listening on port 80",
    "cache cleared for session id xyz-998877",
    "database connected"
]

test_df_spark = spark.createDataFrame([(log,) for log in test_logs], ["text"])
final_predictions = final_pipeline.transform(test_df_spark)

final_predictions.select("text", "prediction", "probability", "rawPrediction").show(truncate=False)

print("=== BATCH PROCESSING EXAMPLE ===")
streaming_df = spark.readStream.format("parquet").load("logs_dataset.parquet")
prediction_stream = final_pipeline.transform(streaming_df)
query = prediction_stream.writeStream \
    .outputMode("complete") \
    .format("console") \
    .trigger(processingTime='10 seconds') \
    .start()

query.awaitTermination(30)
query.stop()

spark.stop()
