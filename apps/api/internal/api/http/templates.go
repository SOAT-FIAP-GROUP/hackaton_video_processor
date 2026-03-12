package http

func GetHTMLForm() string {
	return `<!DOCTYPE html>
<html lang="pt-BR">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>FIAP X - Processador de Vídeos</title>

<style>
body { 
    font-family: Arial, sans-serif; 
    max-width: 800px; 
    margin: 50px auto; 
    padding: 20px;
    background-color: #f5f5f5;
}
.container {
    background: white;
    padding: 30px;
    border-radius: 10px;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}
h1 { 
    color: #333; 
    text-align: center;
    margin-bottom: 30px;
}
.upload-form {
    border: 2px dashed #ddd;
    padding: 30px;
    text-align: center;
    border-radius: 10px;
    margin: 20px 0;
}
input[type="file"] {
    margin: 20px 0;
    padding: 10px;
}
button {
    background: #007bff;
    color: white;
    padding: 12px 30px;
    border: none;
    border-radius: 5px;
    cursor: pointer;
    font-size: 16px;
}
button:hover { background: #0056b3; }

.result {
    margin-top: 20px;
    padding: 15px;
    border-radius: 5px;
    display: none;
}
.success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
.error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }

.loading { 
    text-align: center; 
    display: none;
    margin: 20px 0;
}

.files-list { margin-top: 30px; }

.file-item {
    background: #f8f9fa;
    padding: 10px;
    margin: 5px 0;
    border-radius: 5px;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.download-btn {
    background: #28a745;
    color: white;
    padding: 5px 15px;
    text-decoration: none;
    border-radius: 3px;
    font-size: 14px;
}
.download-btn:hover { background: #218838; }

.logout-btn {
    background: #dc3545;
    color: white;
    padding: 8px 20px;
    border: none;
    border-radius: 5px;
    cursor: pointer;
    font-size: 14px;
    float: right;
}
.logout-btn:hover { background: #c82333; }
</style>
</head>

<body>

<div class="container">

<button class="logout-btn" onclick="logout()">🚪 Logout</button>

<h1>🎬 FIAP X - Processador de Vídeos</h1>

<p style="text-align: center; color: #666;">
Faça upload de um vídeo e receba um ZIP com todos os frames extraídos!
</p>

<form id="uploadForm" class="upload-form">
<p><strong>Selecione um arquivo de vídeo:</strong></p>
<input type="file" id="videoFile" accept="video/*" required>
<br>
<button type="submit">🚀 Processar Vídeo</button>
</form>

<div class="loading" id="loading">
<p>⏳ Processando vídeo... Isso pode levar alguns minutos.</p>
</div>

<div class="result" id="result"></div>

<div class="files-list">
<h3>📁 Arquivos Processados:</h3>
<div id="filesList">Carregando...</div>
</div>

</div>

<script>

document.getElementById('uploadForm').addEventListener('submit', async function(e) {

    e.preventDefault();

    const fileInput = document.getElementById('videoFile');
    const file = fileInput.files[0];

    if (!file) {
        showResult('Selecione um arquivo de vídeo!', 'error');
        return;
    }

    const formData = new FormData();
    formData.append('video', file);

    showLoading(true);
    hideResult();

    try {

        const response = await fetch('/upload', {
            method: 'POST',
            body: formData
        });

        const result = await response.json();

        if (result.success) {

            showResult(
                result.message +
                '<br><br><a href="#" onclick="downloadFile(\'' + result.zip_path + '\')" class="download-btn">⬇️ Download ZIP</a>',
                'success'
            );

            loadFilesList();

        } else {
            showResult('Erro: ' + result.message, 'error');
        }

    } catch (error) {

        showResult('Erro de conexão: ' + error.message, 'error');

    } finally {
        showLoading(false);
    }

});

function showResult(message, type) {
    const result = document.getElementById('result');
    result.innerHTML = message;
    result.className = 'result ' + type;
    result.style.display = 'block';
}

function hideResult() {
    document.getElementById('result').style.display = 'none';
}

function showLoading(show) {
    document.getElementById('loading').style.display = show ? 'block' : 'none';
}

async function downloadFile(filename) {

    try {

        const response = await fetch('/download/' + filename);

        const data = await response.json();

        if (data.download_url) {
            window.location.href = data.download_url;
        }

    } catch (error) {

        console.error("Erro ao baixar arquivo:", error);

    }

}

async function loadFilesList() {

    try {

        const response = await fetch('/api/status');

        const data = await response.json();

        const filesList = document.getElementById('filesList');

        if (data.files && data.files.length > 0) {

            filesList.innerHTML = data.files.map(file => 
                '<div class="file-item">' +
                '<span>' + file.filename + ' (' + formatFileSize(file.size) + ') - ' + file.created_at + '</span>' +
                '<a href="#" onclick="downloadFile(\'' + file.download_url + '\')" class="download-btn">⬇️ Download</a>' +
                '</div>'
            ).join('');

        } else {

            filesList.innerHTML = '<p>Nenhum arquivo processado ainda.</p>';

        }

    } catch (error) {

        document.getElementById('filesList').innerHTML = '<p>Erro ao carregar arquivos.</p>';

    }

}

function formatFileSize(bytes) {

    if (bytes === 0) return '0 Bytes';

    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];

    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];

}

async function logout() {

    try {

        const response = await fetch('/api/logout', { method: 'GET' });

        if (response.ok) {
            window.location.href = '/';
        }

        if (response.redirected) {
            window.location.href = response.url;
        }

    } catch (error) {

        console.error('Erro ao fazer logout:', error);

    }

}

loadFilesList();

</script>

</body>
</html>`
}

func GetLoginPage() string {
	return `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - Sistema</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            max-width: 500px; 
            margin: 50px auto; 
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 { 
            color: #333; 
            text-align: center;
            margin-bottom: 30px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            color: #333;
            margin-bottom: 8px;
            font-weight: bold;
        }
        input[type="email"],
        input[type="password"] {
            width: 100%;
            padding: 12px;
            border: 1px solid #ddd;
            border-radius: 5px;
            font-size: 14px;
            box-sizing: border-box;
        }
        input:focus {
            outline: none;
            border-color: #007bff;
        }
        button {
            width: 100%;
            background: #007bff;
            color: white;
            padding: 12px 30px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
            margin-top: 10px;
        }
        button:hover { 
            background: #0056b3; 
        }
        .result {
            margin-top: 20px;
            padding: 15px;
            border-radius: 5px;
            display: none;
        }
        .success { 
            background: #d4edda; 
            color: #155724; 
            border: 1px solid #c3e6cb; 
        }
        .error { 
            background: #f8d7da; 
            color: #721c24; 
            border: 1px solid #f5c6cb; 
        }
        .links {
            text-align: center;
            margin-top: 20px;
        }
        .links a {
            color: #007bff;
            text-decoration: none;
        }
        .links a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔐 Login</h1>
        
        <form id="loginForm">
            <div class="form-group">
                <label for="email">E-mail:</label>
                <input type="email" id="email" placeholder="seu@email.com" required>
            </div>

            <div class="form-group">
                <label for="password">Senha:</label>
                <input type="password" id="password" placeholder="Digite sua senha" required>
            </div>

            <button type="submit">Entrar</button>
        </form>

        <div class="result" id="result"></div>

        <div class="links">
            <p>Não tem uma conta? <a href="/signup">Cadastre-se aqui</a></p>
        </div>
    </div>

    <script>
        document.getElementById('loginForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const email = document.getElementById('email').value;
            const password = document.getElementById('password').value;
            
            if (!email || !password) {
                showResult('Por favor, preencha todos os campos!', 'error');
                return;
            }
            
            const formData = new FormData();
			formData.append('email', email);
			formData.append('password', password);

			try {
				const response = await fetch('/api/login', {
        						method: 'POST',
        						body: formData
    			});

				if (response.redirected) {
					showResult('✅ Login realizado com sucesso! Redirecionando...', 'success');
					setTimeout(() => {
						window.location.href = response.url;
					}, 1500);
				} else {
					const result = await response.json();
					showResult('Erro: ' + result.message, 'error');
				}
			} catch (err) {
				console.error(err);
				showResult('Erro ao conectar com o servidor', 'error');
			}
        });
        
        function showResult(message, type) {
            const result = document.getElementById('result');
            result.innerHTML = message;
            result.className = 'result ' + type;
            result.style.display = 'block';
        }
    </script>
</body>
</html>`
}

func GetSignupPage() string {
	return `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Cadastro - Sistema</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            max-width: 500px; 
            margin: 50px auto; 
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 { 
            color: #333; 
            text-align: center;
            margin-bottom: 30px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            color: #333;
            margin-bottom: 8px;
            font-weight: bold;
        }
        input[type="text"],
        input[type="email"],
        input[type="password"] {
            width: 100%;
            padding: 12px;
            border: 1px solid #ddd;
            border-radius: 5px;
            font-size: 14px;
            box-sizing: border-box;
        }
        input:focus {
            outline: none;
            border-color: #28a745;
        }
        button {
            width: 100%;
            background: #28a745;
            color: white;
            padding: 12px 30px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
            margin-top: 10px;
        }
        button:hover { 
            background: #218838; 
        }
        .result {
            margin-top: 20px;
            padding: 15px;
            border-radius: 5px;
            display: none;
        }
        .success { 
            background: #d4edda; 
            color: #155724; 
            border: 1px solid #c3e6cb; 
        }
        .error { 
            background: #f8d7da; 
            color: #721c24; 
            border: 1px solid #f5c6cb; 
        }
        .links {
            text-align: center;
            margin-top: 20px;
        }
        .links a {
            color: #28a745;
            text-decoration: none;
        }
        .links a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>📝 Cadastro</h1>
        
        <form id="signupForm">
            <div class="form-group">
                <label for="name">Nome Completo:</label>
                <input type="text" id="name" placeholder="João Silva" required>
            </div>

            <div class="form-group">
                <label for="email">E-mail:</label>
                <input type="email" id="email" placeholder="seu@email.com" required>
            </div>

            <div class="form-group">
                <label for="password">Senha:</label>
                <input type="password" id="password" placeholder="Mínimo 8 caracteres" minlength="8" required>
            </div>

            <div class="form-group">
                <label for="confirmPassword">Confirmar Senha:</label>
                <input type="password" id="confirmPassword" placeholder="Digite a senha novamente" required>
            </div>

            <button type="submit">Cadastrar</button>
        </form>

        <div class="result" id="result"></div>

        <div class="links">
            <p>Já tem uma conta? <a href="/login">Faça login aqui</a></p>
        </div>
    </div>

    <script>
        document.getElementById('signupForm').addEventListener('submit', async function (e) {
    e.preventDefault();

    const name = document.getElementById('name').value.trim();
    const email = document.getElementById('email').value.trim();
    const password = document.getElementById('password').value;
    const confirmPassword = document.getElementById('confirmPassword').value;

    if (!name || !email || !password || !confirmPassword) {
        showResult('Por favor, preencha todos os campos!', 'error');
        return;
    }

    if (password !== confirmPassword) {
        showResult('As senhas não coincidem!', 'error');
        return;
    }

    if (password.length < 8) {
        showResult('A senha deve ter no mínimo 8 caracteres!', 'error');
        return;
    }

		const formData = new FormData();
        formData.append('name', name);
        formData.append('email', email);
        formData.append('password', password);

        try {
            const response = await fetch('/api/signup', {
                method: 'POST',
                body: formData,
				redirect: 'follow'
            });

			if (response.redirected) {
				showResult('✅ Cadastro realizado com sucesso! Redirecionando...', 'success');
				setTimeout(() => {
					window.location.href = response.url;
				}, 1500);
			} else {
				const result = await response.json();
				showResult('Erro: ' + result.message, 'error');
			}

        } catch (err) {
            console.error(err);
            showResult('Erro ao conectar com o servidor', 'error');
        }
});
        
        function showResult(message, type) {
            const result = document.getElementById('result');
            result.innerHTML = message;
            result.className = 'result ' + type;
            result.style.display = 'block';
        }

		function callSignupAPI(name, email, password) {
			
		}
    </script>
</body>
</html>`
}

func Get404Page() string {
	return `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>404 - Página Não Encontrada</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            max-width: 600px; 
            margin: 50px auto; 
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 50px 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            text-align: center;
        }
        .error-code {
            font-size: 100px;
            color: #dc3545;
            font-weight: bold;
            margin: 0;
            line-height: 1;
        }
        h1 { 
            color: #333; 
            margin: 20px 0;
            font-size: 32px;
        }
        p {
            color: #666;
            font-size: 16px;
            line-height: 1.6;
            margin-bottom: 30px;
        }
        .buttons {
            display: flex;
            gap: 15px;
            justify-content: center;
            flex-wrap: wrap;
        }
        .btn {
            padding: 12px 30px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
            text-decoration: none;
            display: inline-block;
        }
        .btn-primary {
            background: #007bff;
            color: white;
        }
        .btn-primary:hover {
            background: #0056b3;
        }
        .btn-secondary {
            background: #28a745;
            color: white;
        }
        .btn-secondary:hover {
            background: #218838;
        }
        .links {
            margin-top: 30px;
            padding-top: 30px;
            border-top: 1px solid #ddd;
        }
        .links h3 {
            color: #333;
            font-size: 18px;
            margin-bottom: 15px;
        }
        .links a {
            color: #007bff;
            text-decoration: none;
            display: block;
            margin: 8px 0;
        }
        .links a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="error-code">404</div>
        <h1>❌ Página Não Encontrada</h1>
        <p>
            Desculpe, a página que você está procurando não existe ou foi removida.<br>
            Verifique o endereço e tente novamente.
        </p>

        <div class="buttons">
            <a href="login.html" class="btn btn-primary">🏠 Ir para Login</a>
            <a href="signup.html" class="btn btn-secondary">📝 Cadastrar</a>
        </div>

        <div class="links">
            <h3>Links Úteis:</h3>
            <a href="#">📚 Central de Ajuda</a>
            <a href="#">💬 Suporte</a>
            <a href="#">📧 Contato</a>
        </div>
    </div>
</body>
</html>`
}
