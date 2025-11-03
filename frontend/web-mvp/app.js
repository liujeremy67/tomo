const selectors = {
  baseUrl: document.querySelector("#base-url"),
  token: document.querySelector("#token"),
  saveSettings: document.querySelector("#save-settings"),
  clearSettings: document.querySelector("#clear-settings"),
  checkAuth: document.querySelector("#check-auth"),
  startTimer: document.querySelector("#start-timer"),
  stopTimer: document.querySelector("#stop-timer"),
  cancelTimer: document.querySelector("#cancel-timer"),
  timerStatus: document.querySelector("#timer-status"),
  timerDisplay: document.querySelector("#timer-display"),
  refreshSessions: document.querySelector("#refresh-sessions"),
  stats: document.querySelector("#stats"),
  sessionList: document.querySelector("#session-list"),
  log: document.querySelector("#log"),
};

const state = {
  baseUrl: "",
  token: "",
  timerStart: null,
  timerInterval: null,
};

class ApiError extends Error {
  constructor(message, { status, body }) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.body = body;
  }
}

const storageKeys = {
  baseUrl: "tomo.mini.baseUrl",
  token: "tomo.mini.token",
};

init();

function init() {
  loadSettings();
  selectors.saveSettings.addEventListener("click", handleSaveSettings);
  selectors.clearSettings.addEventListener("click", handleClearSettings);
  selectors.checkAuth.addEventListener("click", handleCheckAuth);
  selectors.startTimer.addEventListener("click", startTimer);
  selectors.stopTimer.addEventListener("click", stopAndSaveTimer);
  selectors.cancelTimer.addEventListener("click", resetTimer);
  selectors.refreshSessions.addEventListener("click", refreshSessions);

  if (state.token) {
    refreshSessions();
  }
}

function loadSettings() {
  const storedBase = localStorage.getItem(storageKeys.baseUrl);
  const storedToken = localStorage.getItem(storageKeys.token);

  state.baseUrl = storedBase || selectors.baseUrl.placeholder;
  state.token = storedToken || "";

  selectors.baseUrl.value = state.baseUrl;
  selectors.token.value = state.token;
}

function handleSaveSettings() {
  state.baseUrl = selectors.baseUrl.value.trim();
  state.token = selectors.token.value.trim();

  localStorage.setItem(storageKeys.baseUrl, state.baseUrl);
  if (state.token) {
    localStorage.setItem(storageKeys.token, state.token);
  } else {
    localStorage.removeItem(storageKeys.token);
  }

  logMessage("info", "Settings saved.");
}

function handleClearSettings() {
  localStorage.removeItem(storageKeys.baseUrl);
  localStorage.removeItem(storageKeys.token);

  selectors.baseUrl.value = selectors.baseUrl.placeholder;
  selectors.token.value = "";

  state.baseUrl = selectors.baseUrl.placeholder;
  state.token = "";

  logMessage("info", "Settings cleared.");
}

async function handleCheckAuth() {
  try {
    const response = await apiRequest("/me");
    logMessage("success", "GET /me", response);
  } catch (error) {
    logError("GET /me failed", error);
  }
}

function startTimer() {
  if (state.timerStart) {
    return;
  }

  state.timerStart = new Date();
  updateTimerDisplay(0);
  selectors.timerStatus.textContent = `Status: running (${state.timerStart.toLocaleTimeString()})`;

  selectors.startTimer.disabled = true;
  selectors.stopTimer.disabled = false;
  selectors.cancelTimer.disabled = false;

  state.timerInterval = setInterval(() => {
    const elapsed = Date.now() - state.timerStart.getTime();
    updateTimerDisplay(elapsed);
  }, 1000);

  logMessage("info", "Timer started.");
}

async function stopAndSaveTimer() {
  if (!state.timerStart) {
    return;
  }

  if (!state.token) {
    logMessage("warn", "Provide a JWT before saving sessions.");
    return;
  }

  const end = new Date();
  const payload = {
    start_time: state.timerStart.toISOString(),
    end_time: end.toISOString(),
  };

  try {
    const session = await apiRequest("/sessions", {
      method: "POST",
      body: JSON.stringify(payload),
      headers: { "Content-Type": "application/json" },
    });
    logMessage("success", "Session saved.", session);
    resetTimer();
    await refreshSessions();
  } catch (error) {
    logError("Failed to save session", error);
  }
}

function resetTimer() {
  if (state.timerInterval) {
    clearInterval(state.timerInterval);
  }
  state.timerInterval = null;
  state.timerStart = null;
  selectors.timerDisplay.textContent = "00:00:00";
  selectors.timerStatus.textContent = "Status: idle";

  selectors.startTimer.disabled = false;
  selectors.stopTimer.disabled = true;
  selectors.cancelTimer.disabled = true;

  logMessage("info", "Timer reset.");
}

async function refreshSessions() {
  if (!state.token) {
    logMessage("warn", "Provide a JWT to load your sessions.");
    return;
  }

  try {
    const response = await apiRequest("/sessions");
    logMessage("success", "Fetched sessions.", response);
    renderStats(response.stats);
    renderSessions(response.sessions);
  } catch (error) {
    logError("Failed to fetch sessions", error);
  }
}

function renderStats(stats) {
  selectors.stats.innerHTML = "";
  if (!stats) {
    return;
  }

  const entries = [
    ["Total Sessions", stats.session_count],
    ["Total Minutes", stats.total_minutes],
    ["Total Hours", stats.total_hours?.toFixed(2)],
  ];

  entries.forEach(([label, value]) => {
    const pill = document.createElement("div");
    pill.className = "stat-pill";
    pill.innerHTML = `<strong>${label}:</strong> ${value}`;
    selectors.stats.appendChild(pill);
  });
}

function renderSessions(sessions) {
  selectors.sessionList.innerHTML = "";
  if (!Array.isArray(sessions) || sessions.length === 0) {
    const empty = document.createElement("li");
    empty.className = "session-item";
    empty.textContent = "No sessions recorded yet.";
    selectors.sessionList.appendChild(empty);
    return;
  }

  sessions.forEach((session) => {
    const item = document.createElement("li");
    item.className = "session-item";

    const start = new Date(session.start_time);
    const end = new Date(session.end_time);

    item.innerHTML = `
      <time>${start.toLocaleString()} → ${end.toLocaleTimeString()}</time>
      <span>Duration: ${session.duration_minutes} minutes</span>
      <span>Session ID: ${session.id}</span>
    `;

    selectors.sessionList.appendChild(item);
  });
}

async function apiRequest(path, options = {}) {
  const trimmedBase = state.baseUrl?.replace(/\/+$/, "") || "";
  if (!trimmedBase) {
    throw new Error("Set an API base URL first.");
  }

  const url = `${trimmedBase}${path}`;
  const headers = new Headers(options.headers || {});

  if (state.token) {
    headers.set("Authorization", `Bearer ${state.token}`);
  }

  const response = await fetch(url, {
    ...options,
    headers,
  });

  const text = await response.text();
  let data;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = text;
    }
  }

  if (!response.ok) {
    throw new ApiError(`Request failed with ${response.status}`, {
      status: response.status,
      body: data,
    });
  }

  return data;
}

function updateTimerDisplay(milliseconds) {
  const totalSeconds = Math.floor(milliseconds / 1000);
  const hours = Math.floor(totalSeconds / 3600)
    .toString()
    .padStart(2, "0");
  const minutes = Math.floor((totalSeconds % 3600) / 60)
    .toString()
    .padStart(2, "0");
  const seconds = (totalSeconds % 60).toString().padStart(2, "0");

  selectors.timerDisplay.textContent = `${hours}:${minutes}:${seconds}`;
}

function logMessage(level, message, payload) {
  appendLog({
    level,
    message,
    payload,
  });
}

function logError(message, error) {
  const payload =
    error instanceof ApiError
      ? { status: error.status, body: error.body }
      : { message: error.message };

  appendLog({
    level: "error",
    message,
    payload,
  });
}

function appendLog(entry) {
  const timestamp = new Date().toLocaleTimeString();
  const lines = [`[${timestamp}] ${entry.level.toUpperCase()}: ${entry.message}`];

  if (entry.payload !== undefined) {
    lines.push(JSON.stringify(entry.payload, null, 2));
  }

  const block = document.createElement("div");
  block.className = "log-entry";
  block.textContent = lines.join("\n");

  selectors.log.prepend(block);
}
