use std::process::{Child, Command, Stdio};
use tauri::Manager;
use std::sync::Mutex;
use std::time::Duration;

const DEFAULT_PORT: u16 = 38473;

/// Holds the Go backend subprocess. Kills it when dropped.
struct BackendProcess(pub Mutex<Option<Child>>);

impl Drop for BackendProcess {
    fn drop(&mut self) {
        if let Ok(mut guard) = self.0.lock() {
            if let Some(mut child) = guard.take() {
                let _ = child.kill();
            }
        }
    }
}

/// Find the path to the openade-backend binary.
/// Checks: OPENADE_BACKEND_PATH env, then ../backend binary relative to exe, then ./openade-backend
fn find_backend_path() -> Option<std::path::PathBuf> {
    if let Ok(path) = std::env::var("OPENADE_BACKEND_PATH") {
        let p = std::path::PathBuf::from(path);
        if p.exists() {
            return Some(p);
        }
    }

    let exe = std::env::current_exe().ok()?;
    let exe_dir = exe.parent()?;

    // In dev: target/debug/open-ade. Backend might be at ../../../backend (from workspace root)
    // In prod: .app/Contents/MacOS/open-ade. Backend might be in Resources.
    let candidates = [
        exe_dir.join("openade-backend"),
        exe_dir.join("backend"),
        exe_dir.parent()?.join("backend"),
        exe_dir
            .parent()?
            .parent()?
            .parent()?
            .join("backend"),
    ];

    for p in candidates {
        if p.exists() {
            return Some(p);
        }
    }

    None
}

/// Spawn the Go backend. Returns None if binary not found (dev mode: run backend separately).
fn spawn_backend(
    backend_path: &std::path::Path,
    port: u16,
    db_path: &std::path::Path,
) -> std::io::Result<Child> {
    Command::new(backend_path)
        .env("OPENADE_PORT", port.to_string())
        .env("OPENADE_DB_PATH", db_path)
        .stdout(Stdio::null())
        .stderr(Stdio::piped())
        .spawn()
}

/// Wait for backend to respond on /health or /api/health
fn wait_for_backend(port: u16) -> bool {
    let url = format!("http://127.0.0.1:{}/health", port);
    for _ in 0..30 {
        if let Ok(resp) = reqwest::blocking::get(&url) {
            if resp.status().is_success() {
                return true;
            }
        }
        std::thread::sleep(Duration::from_millis(200));
    }
    false
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .setup(|app| {
            let port = std::env::var("OPENADE_PORT")
                .ok()
                .and_then(|s| s.parse().ok())
                .unwrap_or(DEFAULT_PORT);

            let db_path = app.path().app_data_dir().unwrap_or_else(|_| {
                std::path::PathBuf::from(".")
            });
            std::fs::create_dir_all(&db_path).ok();
            let db_file = db_path.join("openade.db");

            if let Some(backend_path) = find_backend_path() {
                match spawn_backend(&backend_path, port, &db_file) {
                    Ok(child) => {
                        app.manage(BackendProcess(Mutex::new(Some(child))));
                        if wait_for_backend(port) {
                            eprintln!("OpenADE backend ready on port {}", port);
                        } else {
                            eprintln!("Backend started but /health not responding. Proceeding anyway.");
                        }
                    }
                    Err(e) => {
                        eprintln!("Failed to spawn backend: {}. Run backend manually.", e);
                    }
                }
            } else {
                eprintln!(
                    "Backend binary not found. Set OPENADE_BACKEND_PATH or run backend manually on port {}.",
                    port
                );
            }

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
