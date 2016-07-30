# wings-task
聚美优品Wing库存系统 下载任务工具包

<h3>AddJob 客户端向任务队列添加任务.</h3>

<code>
    /\*\*
    <p>&nbsp;&nbsp;\* @param string $tube 队列名称.</p>
    <p>&nbsp;&nbsp;\* @param string $data 添加的一个任务的数据.</p>
    <p>&nbsp;&nbsp;\*</p>
    <p>&nbsp;&nbsp;\* @return string|false key 添加任务成功后返回一个唯一KEY.</p>
    <p>&nbsp;&nbsp;**/</p>
    &nbsp;&nbsp;addJob($tube,$data)
</code>

<h3>GetJob Worker端向任务队列获取任务.</h3>

<code>
    /\*\*
    <p>&nbsp;&nbsp;\* @param string $tube 队列名称.</p>
    <p>&nbsp;&nbsp;\* @param string &$_key 成功获取到一个任务,返回的KEY.</p>
    <p>&nbsp;&nbsp;\* @param string &$_data 成功获取到一个任务,返回的数据.</p>
    <p>&nbsp;&nbsp;\*</p>
    <p>&nbsp;&nbsp;\* @return bool 获取任务是否成功</p>
    <p>&nbsp;&nbsp;\*\*/</p>
    &nbsp;&nbsp;GetJob($tube,&$_key, &$_data)
</code>

<h3>GetReturn 客户端获取Worker端完成任务后的结果.</h3>

<code>
    /\*\*
    <p>&nbsp;&nbsp;\* @param string $key 添加任务时,返回的唯一KEY.</p>
    <p>&nbsp;&nbsp;\*</p>
    <p>&nbsp;&nbsp;\* @return string|false 返回结果数据</p>
    <p>&nbsp;&nbsp;\*\*/</p>
    &nbsp;&nbsp;GetReturn($key)
</code>


<h3>SetReturn Worker端完成任务的结果数据.</h3>

<code>
    /\*\*
    <p>&nbsp;&nbsp;\* @param string $key 添加任务时,返回的唯一KEY.</p>
    <p>&nbsp;&nbsp;\* @param string $data 结果数据.</p>
    <p>&nbsp;&nbsp;\*</p>
    <p>&nbsp;&nbsp;\* @return bool 是否设置任务结果成功</p>
    <p>&nbsp;&nbsp;\*\*/</p>
    &nbsp;&nbsp;GetReturn($key)
</code>

<h3>Uer1 Worker向服务端提交一个事件注册,如果队列有新任务则返回.</h3>

<code>
    /\*\*
    <p>&nbsp;&nbsp;\* @param string $tube 队列名称.</p>
    <p>&nbsp;&nbsp;\*</p>
    <p>&nbsp;&nbsp;\* @return bool</p>
    <p>&nbsp;&nbsp;\*\*/</p>
    &nbsp;&nbsp;Uer1($tube)
</code>



