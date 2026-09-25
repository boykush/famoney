-- マネーフォワード ME の CSV（raw）を明細（product）にする。
-- 金額の符号はマネーフォワードのまま: 収入が正、支出が負。
WITH raw AS (
  SELECT * FROM read_csv({{raw}}, header = true, all_varchar = true)
),
typed AS (
  SELECT
    "ID"                                                  AS id,
    strptime("日付", '%Y/%m/%d')::DATE                    AS date,
    "内容"                                                AS description,
    CAST(replace("金額（円）", ',', '') AS BIGINT)        AS amount,
    NULLIF("保有金融機関", '')                            AS account,
    NULLIF("大項目", '')                                  AS category,
    NULLIF("中項目", '')                                  AS subcategory,
    NULLIF("メモ", '')                                    AS memo,
    "計算対象" = '1'                                      AS is_target,
    "振替" = '1'                                          AS is_transfer
  FROM raw
)
SELECT
  *,
  -- 集計で使う区分。振替と計算対象外は収支に入れない。
  CASE
    WHEN is_transfer    THEN 'transfer'
    WHEN NOT is_target  THEN 'excluded'
    WHEN amount >= 0    THEN 'income'
    ELSE 'expense'
  END AS kind
FROM typed
-- 同じ ID が重複して出ることがあるので1件にする。
QUALIFY row_number() OVER (PARTITION BY id ORDER BY date) = 1
ORDER BY date, id
