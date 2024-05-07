function fetchHosts() {
    fetch('/status/hosts')
      .then(response => response.json())
      .then(data => {
        const hosts = data.hosts;
        let html = '<h2>主机列表</h2><div class="list-group">';
        hosts.forEach(host => {
          html += `<a href="#" class="list-group-item list-group-item-action">
                    主机名: ${host.hostname}, IP: ${host.ip}, CPU: ${host.cpu}%, 内存: ${host.memory}GB, 磁盘: ${host.disk}GB
                   </a>`;
        });
        html += '</div>';
        document.getElementById('content').innerHTML = html;
      });
  }
  
  function fetchTasks() {
    fetch('/status/task')
      .then(response => response.json())
      .then(data => {
        let html = '<h2>运行任务</h2><div class="list-group">';
        data.forEach(task => {
          html += `<div class="list-group-item">
                    任务编号: ${task['task number']}, 状态: ${task.status}, 任务名: ${task['task name']}
                    <a href="#" onclick="fetchTaskDetail(${task['task number']})" class="btn btn-primary btn-sm">查看详情</a>
                   </div>`;
        });
        html += '</div>';
        document.getElementById('content').innerHTML = html;
      });
  }
  
  function fetchTaskDetail(number) {
    fetch(`/status/${number}/status`)
      .then(response => response.json())
      .then(data => {
        // 此处处理获取到的任务详细信息
        // 例如弹出模态框显示任务详情
        alert(JSON.stringify(data)); // 示例展示方法
      });
  }