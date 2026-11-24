-- Create "event_log" table
CREATE TABLE "public"."event_log" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "event_name" character varying NOT NULL,
  "user_id" character varying NULL,
  "anon_id" character varying NULL,
  "props" jsonb NULL,
  "channel" character varying NULL,
  "occurred_at" timestamptz NOT NULL,
  "received_at" timestamptz NOT NULL,
  "tenant_id" character varying NOT NULL DEFAULT '',
  PRIMARY KEY ("id")
);
-- Create index "eventlog_created_at" to table: "event_log"
CREATE INDEX "eventlog_created_at" ON "public"."event_log" ("created_at");
-- Create index "eventlog_event_name_occurred_at" to table: "event_log"
CREATE INDEX "eventlog_event_name_occurred_at" ON "public"."event_log" ("event_name", "occurred_at");
-- Create index "eventlog_updated_at" to table: "event_log"
CREATE INDEX "eventlog_updated_at" ON "public"."event_log" ("updated_at");
-- Create index "eventlog_user_id_occurred_at" to table: "event_log"
CREATE INDEX "eventlog_user_id_occurred_at" ON "public"."event_log" ("user_id", "occurred_at");
-- Set comment to column: "event_name" on table: "event_log"
COMMENT ON COLUMN "public"."event_log"."event_name" IS '事件词表名；聚合口径见 module/dashboard 的常量';
-- Set comment to column: "user_id" on table: "event_log"
COMMENT ON COLUMN "public"."event_log"."user_id" IS '登录用户；注册前的匿名行为为 NULL';
-- Set comment to column: "anon_id" on table: "event_log"
COMMENT ON COLUMN "public"."event_log"."anon_id" IS '匿名设备标识；匿名行归因为用户由上游负责';
-- Set comment to column: "props" on table: "event_log"
COMMENT ON COLUMN "public"."event_log"."props" IS '事件附加属性；聚合只按 event_name 计数，不依赖其结构';
-- Set comment to column: "channel" on table: "event_log"
COMMENT ON COLUMN "public"."event_log"."channel" IS '来源渠道；未采集为 NULL';
-- Set comment to column: "occurred_at" on table: "event_log"
COMMENT ON COLUMN "public"."event_log"."occurred_at" IS '事件发生时刻（业务时钟）';
-- Set comment to column: "received_at" on table: "event_log"
COMMENT ON COLUMN "public"."event_log"."received_at" IS '入库时刻（采集时钟）；与 occurred_at 的差值即上报延迟';
-- Set comment to column: "tenant_id" on table: "event_log"
COMMENT ON COLUMN "public"."event_log"."tenant_id" IS '所属租户；匿名事件为空';
-- Create "platform_analytics" table
CREATE TABLE "public"."platform_analytics" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "date" timestamptz NOT NULL,
  "new_registrations" bigint NOT NULL DEFAULT 0,
  "completed_tests" bigint NOT NULL DEFAULT 0,
  "tenant_id" character varying NOT NULL DEFAULT '',
  PRIMARY KEY ("id")
);
-- Create index "platformanalytics_created_at" to table: "platform_analytics"
CREATE INDEX "platformanalytics_created_at" ON "public"."platform_analytics" ("created_at");
-- Create index "platformanalytics_tenant_id_date" to table: "platform_analytics"
CREATE UNIQUE INDEX "platformanalytics_tenant_id_date" ON "public"."platform_analytics" ("tenant_id", "date");
-- Create index "platformanalytics_updated_at" to table: "platform_analytics"
CREATE INDEX "platformanalytics_updated_at" ON "public"."platform_analytics" ("updated_at");
-- Set comment to column: "date" on table: "platform_analytics"
COMMENT ON COLUMN "public"."platform_analytics"."date" IS '统计日；日切后去时刻，写入前统一到租户时区的零点';
-- Set comment to column: "new_registrations" on table: "platform_analytics"
COMMENT ON COLUMN "public"."platform_analytics"."new_registrations" IS '当日新增注册数';
-- Set comment to column: "completed_tests" on table: "platform_analytics"
COMMENT ON COLUMN "public"."platform_analytics"."completed_tests" IS '当日完成测评数';
-- Set comment to column: "tenant_id" on table: "platform_analytics"
COMMENT ON COLUMN "public"."platform_analytics"."tenant_id" IS '所属租户';
-- Create "user_referral" table
CREATE TABLE "public"."user_referral" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "referrer_id" character varying NOT NULL,
  "referrer_name" character varying NOT NULL DEFAULT '',
  "invitee_id" character varying NOT NULL,
  "tenant_id" character varying NOT NULL DEFAULT '',
  PRIMARY KEY ("id")
);
-- Create index "userreferral_created_at" to table: "user_referral"
CREATE INDEX "userreferral_created_at" ON "public"."user_referral" ("created_at");
-- Create index "userreferral_invitee_id" to table: "user_referral"
CREATE UNIQUE INDEX "userreferral_invitee_id" ON "public"."user_referral" ("invitee_id");
-- Create index "userreferral_referrer_id" to table: "user_referral"
CREATE INDEX "userreferral_referrer_id" ON "public"."user_referral" ("referrer_id");
-- Create index "userreferral_updated_at" to table: "user_referral"
CREATE INDEX "userreferral_updated_at" ON "public"."user_referral" ("updated_at");
-- Set comment to column: "referrer_id" on table: "user_referral"
COMMENT ON COLUMN "public"."user_referral"."referrer_id" IS '邀请人；与 system.User 的逻辑引用，不做外键';
-- Set comment to column: "referrer_name" on table: "user_referral"
COMMENT ON COLUMN "public"."user_referral"."referrer_name" IS '邀请时刻的展示名快照；活体身份在 system.User，快照只为让本表独立可渲染';
-- Set comment to column: "invitee_id" on table: "user_referral"
COMMENT ON COLUMN "public"."user_referral"."invitee_id" IS '被邀人；一个用户至多被邀一次';
-- Set comment to column: "tenant_id" on table: "user_referral"
COMMENT ON COLUMN "public"."user_referral"."tenant_id" IS '所属租户';
