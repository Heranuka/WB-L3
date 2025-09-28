CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    message TEXT NOT NULL,
    destination TEXT NOT NULL,
    channel TEXT NOT NULL, -- email, telegram, sms и т.д.
    data_sent_at TIMESTAMPTZ NOT NULL, -- время отправки
    status TEXT NOT NULL, -- created, pending, sent, failed, cancelled
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
    
);