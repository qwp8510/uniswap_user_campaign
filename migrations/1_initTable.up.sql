-- 1_initTable.up.sql

CREATE TABLE "userTask" (
    "id" VARCHAR(32) NOT NULL PRIMARY KEY,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "userAddress" VARCHAR(120) NOT NULL,
    "taskId" VARCHAR(32) NOT NULL,
    "state" VARCHAR(15) NOT NULL
);

CREATE TABLE "userPoint" (
    "id" VARCHAR(32) NOT NULL PRIMARY KEY,
    "userAddress" VARCHAR(120) NOT NULL,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "taskId" VARCHAR(32) NOT NULL,
    "point" INT NOT NULL DEFAULT 0,
    "amount" BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE "task" (
    "id" VARCHAR(32) NOT NULL PRIMARY KEY,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "name" VARCHAR(50) NULL,
    "pairAddress" VARCHAR(120) NULL,
    "startAt" TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE "transaction" (
    "id" VARCHAR(32) NOT NULL PRIMARY KEY,
    "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "blockNum" BIGINT NOT NULL,
    "pairAddress" VARCHAR(120) NOT NULL,
    "senderAddress" VARCHAR(120) NOT NULL,
    "amount0In" NUMERIC(38, 18) NOT NULL DEFAULT 0,
    "amount1In" NUMERIC(38, 18) NOT NULL DEFAULT 0,
    "amount0Out" NUMERIC(38, 18) NOT NULL DEFAULT 0,
    "amount1Out" NUMERIC(38, 18) NOT NULL DEFAULT 0,
    "receiverAddress" VARCHAR(120) NOT NULL
);

CREATE UNIQUE INDEX "unique_pairAddress" ON task("pairAddress");
CREATE UNIQUE INDEX "idx_unique_blocknum_pairaddress" ON transaction ("blockNum", "pairAddress");
CREATE UNIQUE INDEX "idx_unique_userpoint_useraddress_taskid" ON "userPoint" ("userAddress", "taskId");
CREATE UNIQUE INDEX "idx_unique_usertask_useraddress_taskid" ON "userTask" ("userAddress", "taskId");
