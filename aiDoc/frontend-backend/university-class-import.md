# 班级批量导入

- `GET /uni/org/classes/template` 下载 XLSX 模板。列为班级编码、班级名称、所属专业、入学年份和状态。专业可用模板下拉、编码或本校唯一的专业名称。状态可填写启用或停用，留空默认启用。
- `POST /uni/org/classes/import/preview` 与 `POST /uni/org/classes/import/confirm` 使用 `multipart/form-data` 的 `file` 字段。每次最多 5000 条，文件不超过 10 MiB。
- 响应 `data` 含有效行数 `valid`、按行错误 `errors`、预览 `preview` 和是否写入 `imported`。预览列依次是班级编码、名称、专业和入学年份。
- 仅当前高校管理员可调用。班级编码和名称在本校内分别唯一。确认时重新校验，任一行失败则整批不写入，已存在的班级不受影响。
