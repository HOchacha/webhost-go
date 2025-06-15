  #!/bin/bash

  # -----------------------------
  # MariaDB 설치 및 설정 자동화
  # -----------------------------
  DB_NAME="webhost"
  DB_USER="user_service"
  DB_PASS="your_password"
  TABLE_SQL_FILE="init.sql"

  # 1. MariaDB 설치
  echo "📦 Installing MariaDB..."
  sudo apt update && sudo apt install -y mariadb-server

  # 2. MariaDB 서비스 시작
  echo "🚀 Starting MariaDB..."
  sudo systemctl start mariadb
  sudo systemctl enable mariadb

  # 3. 초기 보안 설정
  echo "🔐 Securing MariaDB..."
  sudo mysql -e "DELETE FROM mysql.user WHERE User='';"
  sudo mysql -e "DROP DATABASE IF EXISTS test;"
  sudo mysql -e "DELETE FROM mysql.db WHERE Db='test' OR Db='test\\_%';"
  sudo mysql -e "FLUSH PRIVILEGES;"

  # 4. DB 및 사용자 생성
  echo "🛠️  Creating database and user..."
  sudo mariadb -e "
  CREATE DATABASE IF NOT EXISTS $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
  CREATE USER IF NOT EXISTS '$DB_USER'@'localhost' IDENTIFIED BY '$DB_PASS';
  GRANT ALL PRIVILEGES ON $DB_NAME.* TO '$DB_USER'@'localhost';
  FLUSH PRIVILEGES;
  "

  # 5. 테이블 스키마 생성 (init.sql 적용)
  if [ -f "$TABLE_SQL_FILE" ]; then
    echo "📄 Applying schema from $TABLE_SQL_FILE..."
    sudo mariadb "$DB_NAME" < "$TABLE_SQL_FILE"
  else
    echo "⚠️  $TABLE_SQL_FILE not found. Skipping table creation."
  fi

  echo "✅ Done! MariaDB is ready."
