# 补全前置处理详细内容

## 4. 补全前置处理（第494-511行）

### 4.1 调用 prepare_prompt() 方法进行prompt拼接策略处理

**代码位置：** 第494-497行

```python
# 补全前置处理 （拼接prompt策略，单行/多行补全策略）
prompt, prompt_tokens, new_prefix, new_suffix = self.prepare_prompt(prefix, suffix, code_context)
data['prompt'] = prompt
data['prompt_tokens'] = prompt_tokens
```

**功能说明：**
- 调用 `prepare_prompt()` 方法对prefix、suffix和code_context进行处理
- 生成最终的prompt字符串和token数量
- 将处理后的prompt和token数量保存到data字典中
- 返回处理后的new_prefix和new_suffix

### 4.2 判断单行/多行补全类型

**代码位置：** 第499-500行

```python
is_single_completion = completion_line_handler.judge_single_completion(
    cursor_line_prefix, cursor_line_suffix, language)
```

**功能说明：**
- 调用 `completion_line_handler.judge_single_completion()` 方法
- 传入cursor_line_prefix、cursor_line_suffix和language参数
- 判断当前补全请求是单行补全还是多行补全
- 返回布尔值：True表示单行补全，False表示多行补全

### 4.3 根据补全类型调整stop_words

**代码位置：** 第501-506行

```python
if is_single_completion:
    if isinstance(data["stop"], list):
        data["stop"].append("\n")
    elif isinstance(data["stop"], str):
        data["stop"] += "\n"
```

**功能说明：**
- 如果是单行补全（is_single_completion为True）
- 检查data["stop"]的类型：
  - 如果是列表类型，在列表末尾添加换行符"\n"
  - 如果是字符串类型，在字符串末尾添加换行符"\n"
- 这样可以确保单行补全在遇到换行符时停止生成

### 4.4 记录处理耗时和相关信息

**代码位置：** 第507-511行

```python
logger.info(f"前置准备模块耗时：{(time.time() - st) * 1000: .4f}ms, "
            f"language={language}, "
            f"is_single_completion={is_single_completion}, "
            f"prompt_tokens={prompt_tokens}, "
            f"trigger_mode={trigger_mode}", request_id=self.request_id)
```

**功能说明：**
- 计算前置准备模块的总耗时（从开始时间st到当前时间）
- 记录关键信息到日志：
  - 耗时（毫秒）
  - 编程语言
  - 是否单行补全
  - prompt的token数量
  - 触发模式
- 使用request_id进行日志追踪

---

## 相关方法说明

### prepare_prompt() 方法（第344-359行）
```python
def prepare_prompt(self, prefix, suffix, code_context=""):
    """
     前后缀长度处理后对prompt进行拼接
    :param prefix:
    :param suffix:
    :param code_context:
    :return:
    """
    prefix, code_context = self.handle_prompt(prompt=prefix, is_prefix=True,
                                              optional_prompt=code_context,
                                              min_prompt_token=self.min_prefix_token)
    suffix, _ = self.handle_prompt(prompt=suffix, is_prefix=False)
    prompt = self.get_prompt_template(prefix, suffix, code_context)
    prompt_tokens = self.get_token_num(prompt)
    self.is_windows = WIN_NL in prefix
    return prompt, prompt_tokens, prefix, suffix
```

**功能说明：**
- 处理prefix和code_context的token长度限制
- 处理suffix的token长度限制
- 使用get_prompt_template()方法拼接最终的prompt
- 计算prompt的token数量
- 判断是否为Windows系统（根据换行符）

### judge_single_completion() 方法
此方法位于 `CompletionLineHandler` 类中，用于判断补全类型：
- 分析cursor_line_prefix和cursor_line_suffix
- 结合编程语言特性
- 返回补全类型判断结果