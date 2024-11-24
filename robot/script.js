import http from 'k6/http';
import { check, sleep, current } from 'k6';

// API 的基础 URL
const baseUrl = 'http://127.0.0.1:8080/api/v1';

// 颜色转义码
const colors = {
    green: '\x1b[32m',  // 绿色
    yellow: '\x1b[33m', // 黄色
    reset: '\x1b[0m'    // 重置颜色
};

// 常量定义
const roundInterval = 5;
const clickCount = 100;
const robotId = 10001;
const battleId = 1;

// k6 的负载测试选项
export const options = {
    vus: 1000,  // 虚拟用户数量
    duration: '1h',  // 测试持续时间
};

// 打印日志
function logRequest(userId, url, data, isRequest = true) {
    console.log(
        `${colors.yellow}${userId}${colors.reset} ${isRequest ? '===>' : '<==='} ${colors.green}${url}${colors.reset}`,
        data || ""
    );
}

// 执行 HTTP POST 请求的函数
function request(url, token, data, code) {
    const refinedParams = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
        },
    };
    const payload = data ? JSON.stringify(data) : ""; // 将数据对象转为 JSON 字符串
    const userId = getUUID(); // 当前虚拟用户 ID

    logRequest(userId, url, data); // 打印请求日志
    const res = http.post(baseUrl + url, payload, refinedParams); // 发送 HTTP POST 请求
    logRequest(userId, url, res.body, false); // 打印响应日志

    const resp = JSON.parse(res.body); // 解析响应 JSON 数据
    checkAndThrow(resp, url, code);
    return resp;
}

// 检查数据并在检查失败时抛出错误的函数
function checkAndThrow(data, checkName, code) {
    const passed = check(data, {
        [checkName]: (r) => {
            return code.indexOf(r.code) !== -1;
        }
        , // 检查响应代码是否为 0
    });
    if (!passed) {
        throw new Error(`Failed check: ${checkName}`); // 抛出检查失败的错误
    }
}

// 获取当前虚拟用户的运行时间
function currentVUTime() {
    return Math.floor(new Date().getTime() / 1000);
}

// 机器人uuid
function getUUID() {
    return `robot-${robotId + __VU}`;
}

// 战斗逻辑
function battleLogic(token, timeOffset, deskId) {
    let resp = request("/battleroundresult", token, { deskId }, [0]);
    if (resp.data && resp.data.state === 1) {
        resp = request("/battlesettlement", token, { deskId }, [0]);
        return; // 退出战斗逻辑
    }

    let { roundStartTime, state } = resp.data;
    let { grid } = resp.data.syncScore;
    let curRound = 0;
    let scoreSyncCounter = 0;

    while (true) {
        sleep(1); // 每次循环暂停 1 秒

        const now = currentVUTime() + timeOffset; // 当前时间加上时间偏移

        if (roundStartTime - now > 0) {
            if (grid === -1 && state != 3 && state != 1) {
                const gridList = [1, 2, 3, 4, 5, 6, 7, 8];
                if (resp.data.result) {
                    for (let k = 0; k < resp.data.result.length; k++) {
                        const index = gridList.indexOf(resp.data.result[k]);
                        if (index != -1) {
                            gridList.splice(index, 1);
                        }
                    }
                }
                const index = Math.floor(Math.random() * gridList.length);
                resp = request("/battlebet", token, { deskId, grid: gridList[index] }, [0]);
                if (resp.code == 0) {
                    grid = gridList[index];
                }
            }
        } else {
            resp = request("/battleroundresult", token, { deskId }, [0]);
            if (!resp.data.result || resp.data.result.length < resp.data.round || resp.data.round === curRound) {
                if (resp.data.state === 1) { // 直接结算
                    resp = request("/battlesettlement", token, { deskId }, [0]);
                    return; // 退出战斗逻辑
                }
                roundStartTime = resp.data.roundStartTime;
                scoreSyncCounter = 0;
                grid = resp.data.syncScore.grid;
            } else {
                const roundTime = roundStartTime + roundInterval - now;
                if (roundTime > 0) {
                    sleep(roundInterval);
                    curRound = resp.data.round;

                    if (resp.data.state === 1 || resp.data.state === 3) { // 准备结算
                        resp = request("/battlesettlement", token, { deskId }, [0]);
                        return; // 退出战斗逻辑
                    } else if (resp.data.state === 0) {
                        resp = request("/battleroundresult", token, { deskId }, [0]);
                        if (resp.data.state === 1) { // 直接结算
                            resp = request("/battlesettlement", token, { deskId }, [0]);
                            return; // 退出战斗逻辑
                        } else {
                            roundStartTime = resp.data.roundStartTime;
                            scoreSyncCounter = 0;
                            grid = resp.data.syncScore.grid;
                        }
                    }
                } else {
                    resp = request("/battleroundresult", token, { deskId }, [0]);
                    if (resp.data.state === 1) { // 直接结算
                        resp = request("/battlesettlement", token, { deskId }, [0]);
                        return; // 退出战斗逻辑
                    } else {
                        roundStartTime = resp.data.roundStartTime;
                        scoreSyncCounter = 0;
                        grid = resp.data.syncScore.grid;
                    }
                }
            }
        }

        if (state === 0) {
            scoreSyncCounter++;
            if (scoreSyncCounter >= 5) {
                scoreSyncCounter = 0;
                request("/battlesyncscore", token, { deskId }, [0]);
            }
        }
    }
}


// 默认函数，定义 VU 的逻辑
export default function () {
    const userId = getUUID();
    console.log(`${colors.yellow}${userId}${colors.reset}`);

    let token = "";
    let startTime = currentVUTime(); // 记录开始时间

    let resp = request("/login", token, { locale: "", userUid: userId, username: "", startParam: "" }, [0]);
    token = resp.data.token;
    sleep(1);

    resp = request("/enter", token, {}, [0]);
    token = resp.data.token;
    sleep(1);

    const timeOffset = resp.data.serverTime - currentVUTime(); // 计算时间偏移
    const lastDeskId = resp.data.lastDeskId;

    if (lastDeskId) {
        battleLogic(token, timeOffset, lastDeskId); // 继续战斗逻辑
    } else {
        let petId = 0;
        const petData = resp.data.pets[0];

        if (petData) {
            petId = petData.petId; // 如果有宠物，获取宠物 ID
        } else {
            // 如果没有宠物，通过点击屏幕获得宠物
            while (true) {
                resp = request("/clickscreen", token, { clickCount }, [0]);
                sleep(0.5);

                if (resp.data && resp.data.reward && resp.data.reward.egg) {
                    resp = request("/eggopen", token, { id: resp.data.reward.egg.id }, [0]);
                    if (resp.data && resp.data.item) {
                        petId = resp.data.item.id; // 获取宠物 ID
                        break;
                    }
                    sleep(0.5);
                }
            }
        }

        let roomData = {};

        while (true) {
            sleep(1);

            resp = request("/battlematch", token, { battleId, petId }, [0, 21, 27, 25, 26, -3]);
            if (resp.code == -3) {
                sleep(Math.random() * 3);
            } else if (resp.code == 0) {
                roomData.deskId = resp.data.matchState.deskId;
                roomData.createdAt = resp.data.matchState.createdAt;
                break
            } else if (resp.code == 27) { // 报名人数已满
                console.log(`${colors.yellow}User ${userId} Waiting for the next battle.${colors.reset}`);
            }
        }

        // 等待战斗开始
        while (true) {
            sleep(1);
            resp = request("/battlematchstate", token, { deskId: roomData.deskId }, [0, 21, 24, -3]);
            if (resp.data && resp.data.startAt != 0) {
                break
            }
        }

        battleLogic(token, timeOffset, roomData.deskId); // 继续战斗逻辑
    }

    // 确保战斗逻辑完全执行完毕后才能退出
    const elapsed = currentVUTime() - startTime;
    console.log(`${colors.yellow}User ${userId} finished battle logic. Elapsed time: ${elapsed}s.${colors.reset}`);
}
