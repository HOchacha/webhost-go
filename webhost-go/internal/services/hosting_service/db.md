```sql
CREATE TABLE hosting_plans (
    name VARCHAR(50) PRIMARY KEY,        -- "small", "medium", "large"
    cpu INT NOT NULL,
    memory_mb INT NOT NULL,
    disk_gb INT NOT NULL
);
```


```sql
INSERT INTO hosting_plans (name, cpu, memory_mb, disk_gb) VALUES
('small', 1, 1024, 10),
('medium', 2, 2048, 20),
('large', 4, 4096, 40);
```



```sql
CREATE TABLE hostings (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,               -- libvirt domain name
    plan VARCHAR(50) NOT NULL,                -- references hosting_plans.name
    ip_address VARCHAR(50),                   -- VM 내부 IP (libvirt-agent로부터 수신)
    ssh_port INT,                             -- 외부에서 연결되는 SSH 포트 (nginx-agent로 매핑)
    status VARCHAR(20) DEFAULT 'Pending',     -- Running / Stopped / Error / Pending 등
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (plan) REFERENCES hosting_plans(name)
);
```