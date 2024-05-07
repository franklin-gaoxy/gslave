document.getElementById('loginForm').addEventListener('submit', function(e) {
  e.preventDefault();

  // 获取用户输入的用户名和密码
  const username = document.getElementById('username').value;
  const password = document.getElementById('password').value;

  // 发送POST请求到/login/接口
  fetch('/login/', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      username: username,
      password: password,
    }),
  })
  .then(response => response.json())
  .then(data => {
    if (data.status) {
      // 登录成功，保存token并跳转到home.html
      localStorage.setItem('token', data.token);
      window.location.href = '/home.html';
    } else {
      // 登录失败，显示错误消息
      document.getElementById('loginError').style.display = 'block';
    }
  })
  .catch((error) => {
    console.error('Error:', error);
    document.getElementById('loginError').style.display = 'block';
  });
});