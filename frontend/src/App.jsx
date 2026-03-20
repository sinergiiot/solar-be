import { useEffect, useRef, useState } from "react";
import { FiAlertCircle, FiBarChart2, FiCheckCircle, FiCpu, FiHome, FiLogOut, FiMenu, FiSettings, FiSun, FiUser, FiX } from "react-icons/fi";

import {
  createSolarProfile,
  createDevice,
  deleteDevice,
  deleteSolarProfile,
  getActualHistory,
  getDeviceHeartbeatSummary,
  getDashboardSummary,
  getForecastHistory,
  getMe,
  getMySolarProfiles,
  getNotificationPreferences,
  listDevices,
  rotateDeviceKey,
  updateDevice,
  updateSolarProfile,
  getStoredRefreshToken,
  getStoredToken,
  getTodayForecast,
  login,
  logout,
  recordActualDaily,
  register,
  setStoredRefreshToken,
  setStoredToken,
  updateNotificationPreferences,
} from "./api";
import { getDateDaysAgo, getHistoryRowKey, getHourlyDistribution, getTodayLocalDate, getWeatherRisk } from "./utils";
import AccountSection from "./components/AccountSection";
import AuthFlow from "./components/AuthFlow";
import DashboardSection from "./components/DashboardSection";
import ForecastSection from "./components/ForecastSection";
import HistorySection from "./components/HistorySection";
import IntegrationSection from "./components/IntegrationSection";
import ProfileSection from "./components/ProfileSection";

const emptyRegisterForm = {
  name: "",
  email: "",
  password: "",
};

const emptyLoginForm = {
  email: "",
  password: "",
};

const emptyProfileForm = {
  site_name: "",
  capacity_kwp: "",
  lat: "",
  lng: "",
  tilt: "",
  azimuth: "",
};

const emptyActualForm = {
  date: getTodayLocalDate(),
  actual_kwh: "",
  source: "manual",
};

const emptyDeviceForm = {
  name: "",
  external_id: "",
  solar_profile_id: "",
};

const emptyNotificationPreference = {
  plan_tier: "free",
  primary_channel: "email",
  email_enabled: true,
  telegram_enabled: false,
  whatsapp_enabled: false,
  telegram_chat_id: "",
  whatsapp_phone_e164: "",
  whatsapp_opted_in: false,
  timezone: "Asia/Jakarta",
  preferred_send_time: "06:00:00",
};

// App renders auth and dashboard flows with user-separated data.
export default function App() {
  const [token, setToken] = useState(getStoredToken());
  const [currentUser, setCurrentUser] = useState(null);
  const [profile, setProfile] = useState(null);
  const [profiles, setProfiles] = useState([]);
  const [forecast, setForecast] = useState(null);
  const [summary, setSummary] = useState(null);
  const [todayForecastDashboard, setTodayForecastDashboard] = useState(null);
  const [dashboardForecastHistory, setDashboardForecastHistory] = useState([]);
  const [dashboardActualHistory, setDashboardActualHistory] = useState([]);
  const [forecastHistory, setForecastHistory] = useState([]);
  const [actualHistory, setActualHistory] = useState([]);
  const [selectedForecastProfileID, setSelectedForecastProfileID] = useState("");
  const [historyProfileID, setHistoryProfileID] = useState("");
  const [historyStartDate, setHistoryStartDate] = useState(getDateDaysAgo(30));
  const [historyEndDate, setHistoryEndDate] = useState(getTodayLocalDate());
  const [devices, setDevices] = useState([]);
  const [heartbeatSummary, setHeartbeatSummary] = useState(null);
  const [deviceForm, setDeviceForm] = useState(emptyDeviceForm);
  const [latestDeviceKey, setLatestDeviceKey] = useState("");

  const [registerForm, setRegisterForm] = useState(emptyRegisterForm);
  const [loginForm, setLoginForm] = useState(emptyLoginForm);
  const [profileForm, setProfileForm] = useState(emptyProfileForm);
  const [actualForm, setActualForm] = useState(emptyActualForm);
  const [editingProfileID, setEditingProfileID] = useState("");
  const [editingDeviceID, setEditingDeviceID] = useState("");
  const [editingDeviceIsActive, setEditingDeviceIsActive] = useState(true);

  const [authPage, setAuthPage] = useState("landing");
  const [simCapacity, setSimCapacity] = useState(10);
  const [simCloudCover, setSimCloudCover] = useState(30);
  const [simTimePreset, setSimTimePreset] = useState("noon");
  const [electricityTariff, setElectricityTariff] = useState(1444);
  const [selectedHistoryRowKey, setSelectedHistoryRowKey] = useState("");
  const [notificationPreference, setNotificationPreference] = useState(emptyNotificationPreference);
  const [isLoadingNotificationPreference, setIsLoadingNotificationPreference] = useState(false);
  const [isSavingNotificationPreference, setIsSavingNotificationPreference] = useState(false);

  const [isLoading, setIsLoading] = useState(false);
  const [isAuthLoading, setIsAuthLoading] = useState(false);
  const [isSavingProfile, setIsSavingProfile] = useState(false);
  const [isLoadingForecast, setIsLoadingForecast] = useState(false);
  const [isSavingActual, setIsSavingActual] = useState(false);
  const [isLocating, setIsLocating] = useState(false);
  const [isLoadingDevices, setIsLoadingDevices] = useState(false);
  const [isLoadingHistory, setIsLoadingHistory] = useState(false);
  const [isSavingDevice, setIsSavingDevice] = useState(false);
  const [deletingProfileID, setDeletingProfileID] = useState("");
  const [deletingDeviceID, setDeletingDeviceID] = useState("");
  const [rotatingDeviceID, setRotatingDeviceID] = useState("");

  const [feedback, setFeedback] = useState("");
  const [feedbackFading, setFeedbackFading] = useState(false);
  const [error, setError] = useState("");
  const [currentPage, setCurrentPage] = useState("dashboard");
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const feedbackTimeoutRef = useRef(null);
  const feedbackFadeRef = useRef(null);

  useEffect(() => {
    if (!token) {
      setCurrentUser(null);
      setProfile(null);
      setForecast(null);
      return;
    }

    loadAuthenticatedDashboard();
  }, [token]);

  useEffect(() => {
    if (!token || currentPage !== "integration") {
      return;
    }

    loadDevices();
  }, [token, currentPage]);

  useEffect(() => {
    if (!token || currentPage !== "account") {
      return;
    }

    loadNotificationPreference();
  }, [token, currentPage]);

  useEffect(() => {
    if (!token) {
      setTodayForecastDashboard(null);
      return;
    }

    loadTodayForecastDashboard();
  }, [token, selectedForecastProfileID]);

  useEffect(() => {
    if (!token) {
      return undefined;
    }

    const now = new Date();
    const nextSixAM = new Date(now);
    nextSixAM.setHours(6, 0, 0, 0);
    if (now >= nextSixAM) {
      nextSixAM.setDate(nextSixAM.getDate() + 1);
    }

    const msUntilSixAM = nextSixAM.getTime() - now.getTime();
    let intervalID = null;

    const timeoutID = window.setTimeout(() => {
      loadTodayForecastDashboard();
      intervalID = window.setInterval(loadTodayForecastDashboard, 24 * 60 * 60 * 1000);
    }, msUntilSixAM);

    return () => {
      window.clearTimeout(timeoutID);
      if (intervalID) {
        window.clearInterval(intervalID);
      }
    };
  }, [token, selectedForecastProfileID]);

  useEffect(() => {
    if (!feedback) {
      clearTimeout(feedbackTimeoutRef.current);
      clearTimeout(feedbackFadeRef.current);
      setFeedbackFading(false);
      return undefined;
    }

    setFeedbackFading(false);
    feedbackFadeRef.current = setTimeout(() => setFeedbackFading(true), 4700);
    feedbackTimeoutRef.current = setTimeout(() => {
      setFeedback("");
      setFeedbackFading(false);
    }, 5000);

    return () => {
      clearTimeout(feedbackTimeoutRef.current);
      clearTimeout(feedbackFadeRef.current);
    };
  }, [feedback]);

  // Pre-fill form when profile is loaded
  useEffect(() => {
    if (profile) {
      setProfileForm({
        site_name: profile.site_name || "",
        capacity_kwp: profile.capacity_kwp?.toString() || "",
        lat: profile.lat?.toString() || "",
        lng: profile.lng?.toString() || "",
        tilt: profile.tilt?.toString() || "",
        azimuth: profile.azimuth?.toString() || "",
      });

      setDeviceForm((current) => ({
        ...current,
        solar_profile_id: profile.id || "",
      }));
    }
  }, [profile]);

  async function loadAuthenticatedDashboard() {
    setIsLoading(true);
    setError("");

    try {
      const me = await getMe();
      setCurrentUser(me);

      try {
        const profilesData = await getMySolarProfiles();
        const nextProfiles = profilesData.profiles || [];
        setProfiles(nextProfiles);
        setProfile(nextProfiles.length > 0 ? nextProfiles[0] : null);
        setSelectedForecastProfileID(nextProfiles[0]?.id || "");
        setHistoryProfileID("");
      } catch {
        setProfiles([]);
        setProfile(null);
        setSelectedForecastProfileID("");
        setHistoryProfileID("");
      }

      // Load summary and dashboard snapshot data (fixed scope)
      try {
        const summaryData = await getDashboardSummary();
        setSummary(summaryData);
      } catch {
        setSummary(null);
      }

      await loadDashboardSnapshotData();

      await loadHistoryData({
        profile_id: "",
        start_date: getDateDaysAgo(30),
        end_date: getTodayLocalDate(),
      });

      try {
        const heartbeat = await getDeviceHeartbeatSummary();
        setHeartbeatSummary(heartbeat);
      } catch {
        setHeartbeatSummary(null);
      }
    } catch (loadError) {
      setError(loadError.message);
      handleClientLogout(false);
    } finally {
      setIsLoading(false);
    }
  }

  // loadDashboardSnapshotData fetches fixed 7-day snapshot for dashboard cards/table.
  async function loadDashboardSnapshotData() {
    const snapshotFilter = {
      profile_id: undefined,
      start_date: getDateDaysAgo(7),
      end_date: getTodayLocalDate(),
    };

    try {
      const [historyData, actualsData] = await Promise.all([getForecastHistory(snapshotFilter), getActualHistory(snapshotFilter)]);
      setDashboardForecastHistory(historyData.forecasts || []);
      setDashboardActualHistory(actualsData.actuals || []);
    } catch {
      setDashboardForecastHistory([]);
      setDashboardActualHistory([]);
    }
  }

  // loadTodayForecastDashboard fetches today's forecast for dashboard narrative cards.
  async function loadTodayForecastDashboard() {
    try {
      const forecastData = await getTodayForecast({
        profile_id: selectedForecastProfileID || undefined,
      });
      setTodayForecastDashboard(forecastData);
    } catch {
      setTodayForecastDashboard(null);
    }
  }

  function handleClientLogout(callApi = true) {
    if (callApi) {
      logout().catch(() => null);
    }

    setStoredToken("");
    setStoredRefreshToken("");
    setToken("");
    setCurrentUser(null);
    setProfile(null);
    setProfiles([]);
    setForecast(null);
    setSummary(null);
    setTodayForecastDashboard(null);
    setDashboardForecastHistory([]);
    setDashboardActualHistory([]);
    setForecastHistory([]);
    setActualHistory([]);
    setSelectedForecastProfileID("");
    setHistoryProfileID("");
    setHistoryStartDate(getDateDaysAgo(30));
    setHistoryEndDate(getTodayLocalDate());
    setHeartbeatSummary(null);
    setDevices([]);
    setDeviceForm(emptyDeviceForm);
    setLatestDeviceKey("");
    setRegisterForm(emptyRegisterForm);
    setLoginForm(emptyLoginForm);
    setProfileForm(emptyProfileForm);
    setActualForm(emptyActualForm);
    setEditingProfileID("");
    setEditingDeviceID("");
    setEditingDeviceIsActive(true);
    setDeletingProfileID("");
    setDeletingDeviceID("");
    setNotificationPreference(emptyNotificationPreference);
    setFeedback("");
    setError("");
    setAuthPage("landing");
  }

  async function loadNotificationPreference() {
    setIsLoadingNotificationPreference(true);

    try {
      const result = await getNotificationPreferences();
      setNotificationPreference({
        plan_tier: result.plan_tier || "free",
        primary_channel: result.primary_channel || "email",
        email_enabled: Boolean(result.email_enabled),
        telegram_enabled: Boolean(result.telegram_enabled),
        whatsapp_enabled: Boolean(result.whatsapp_enabled),
        telegram_chat_id: result.telegram_chat_id || "",
        whatsapp_phone_e164: result.whatsapp_phone_e164 || "",
        whatsapp_opted_in: Boolean(result.whatsapp_opted_in),
        timezone: result.timezone || "Asia/Jakarta",
        preferred_send_time: result.preferred_send_time || "06:00:00",
      });
    } catch (loadError) {
      setError(loadError.message);
    } finally {
      setIsLoadingNotificationPreference(false);
    }
  }

  async function handleSaveNotificationPreference(event) {
    event.preventDefault();
    setIsSavingNotificationPreference(true);
    setError("");
    setFeedback("");

    try {
      const result = await updateNotificationPreferences(notificationPreference);
      setNotificationPreference({
        plan_tier: result.plan_tier || "free",
        primary_channel: result.primary_channel || "email",
        email_enabled: Boolean(result.email_enabled),
        telegram_enabled: Boolean(result.telegram_enabled),
        whatsapp_enabled: Boolean(result.whatsapp_enabled),
        telegram_chat_id: result.telegram_chat_id || "",
        whatsapp_phone_e164: result.whatsapp_phone_e164 || "",
        whatsapp_opted_in: Boolean(result.whatsapp_opted_in),
        timezone: result.timezone || "Asia/Jakarta",
        preferred_send_time: result.preferred_send_time || "06:00:00",
      });
      setFeedback("Pengaturan notifikasi berhasil disimpan.");
    } catch (saveError) {
      setError(saveError.message);
    } finally {
      setIsSavingNotificationPreference(false);
    }
  }

  async function handleRegister(event) {
    event.preventDefault();
    setIsAuthLoading(true);
    setError("");
    setFeedback("");

    try {
      const payload = await register(registerForm);
      setStoredToken(payload.access_token);
      setStoredRefreshToken(payload.refresh_token);
      setToken(payload.access_token);
      setRegisterForm(emptyRegisterForm);
      setFeedback(`Akun ${payload.user.name} berhasil dibuat dan login.`);
    } catch (authError) {
      setError(authError.message);
    } finally {
      setIsAuthLoading(false);
    }
  }

  async function handleLogin(event) {
    event.preventDefault();
    setIsAuthLoading(true);
    setError("");
    setFeedback("");

    try {
      const payload = await login(loginForm);
      setStoredToken(payload.access_token);
      setStoredRefreshToken(payload.refresh_token);
      setToken(payload.access_token);
      setLoginForm(emptyLoginForm);
      setFeedback(`Login berhasil. Selamat datang ${payload.user.name}.`);
    } catch (authError) {
      setError(authError.message);
    } finally {
      setIsAuthLoading(false);
    }
  }

  async function handleCreateProfile(event) {
    event.preventDefault();
    setIsSavingProfile(true);
    setError("");
    setFeedback("");

    try {
      const payload = {
        site_name: profileForm.site_name,
        capacity_kwp: Number(profileForm.capacity_kwp),
        lat: Number(profileForm.lat),
        lng: Number(profileForm.lng),
        tilt: profileForm.tilt === "" ? null : Number(profileForm.tilt),
        azimuth: profileForm.azimuth === "" ? null : Number(profileForm.azimuth),
      };

      if (editingProfileID) {
        await updateSolarProfile(editingProfileID, payload);
      } else {
        await createSolarProfile(payload);
      }

      const profilesData = await getMySolarProfiles();
      const nextProfiles = profilesData.profiles || [];
      setProfiles(nextProfiles);
      setProfile(nextProfiles.length > 0 ? nextProfiles[0] : null);
      setFeedback(editingProfileID ? "Solar profile berhasil diperbarui." : "Solar profile berhasil disimpan untuk akun ini.");
      setProfileForm({ ...emptyProfileForm, site_name: "Main Site" });
      setEditingProfileID("");
    } catch (submitError) {
      setError(submitError.message);
    } finally {
      setIsSavingProfile(false);
    }
  }

  async function handleDeleteProfile(profileID) {
    if (!window.confirm("Hapus solar profile ini? Device akan terlepas dari profile ini.")) {
      return;
    }

    setDeletingProfileID(profileID);
    setError("");
    setFeedback("");

    try {
      await deleteSolarProfile(profileID);
      const profilesData = await getMySolarProfiles();
      const nextProfiles = profilesData.profiles || [];
      setProfiles(nextProfiles);
      const nextCurrentProfile = nextProfiles.length > 0 ? nextProfiles[0] : null;
      setProfile(nextCurrentProfile);
      setSelectedForecastProfileID(nextCurrentProfile?.id || "");
      if (editingProfileID === profileID) {
        setEditingProfileID("");
        setProfileForm({ ...emptyProfileForm, site_name: "Main Site" });
      }
      setFeedback("Solar profile berhasil dihapus.");
    } catch (deleteError) {
      setError(deleteError.message);
    } finally {
      setDeletingProfileID("");
    }
  }

  function handleEditProfile(profileItem) {
    setEditingProfileID(profileItem.id);
    setProfileForm({
      site_name: profileItem.site_name || "",
      capacity_kwp: profileItem.capacity_kwp?.toString() || "",
      lat: profileItem.lat?.toString() || "",
      lng: profileItem.lng?.toString() || "",
      tilt: profileItem.tilt?.toString() || "",
      azimuth: profileItem.azimuth?.toString() || "",
    });
    setCurrentPage("profile");
  }

  function handleCancelProfileEdit() {
    setEditingProfileID("");
    setProfileForm({ ...emptyProfileForm, site_name: "Main Site" });
  }

  async function handleLoadForecast() {
    setIsLoadingForecast(true);
    setError("");
    setFeedback("");

    try {
      const nextForecast = await getTodayForecast({
        profile_id: selectedForecastProfileID || undefined,
      });
      setForecast(nextForecast);
      setFeedback("Forecast harian berhasil dimuat.");
    } catch (forecastError) {
      setError(forecastError.message);
      setForecast(null);
    } finally {
      setIsLoadingForecast(false);
    }
  }

  async function handleRecordActual(event) {
    event.preventDefault();

    setIsSavingActual(true);
    setError("");
    setFeedback("");

    try {
      await recordActualDaily({
        solar_profile_id: selectedForecastProfileID || undefined,
        date: actualForm.date,
        actual_kwh: Number(actualForm.actual_kwh),
        source: actualForm.source,
      });

      setFeedback("Actual harian berhasil disimpan untuk akun ini.");
      setActualForm((current) => ({ ...current, actual_kwh: "" }));
      await loadDashboardSnapshotData();
      await loadHistoryData();
    } catch (submitError) {
      setError(submitError.message);
    } finally {
      setIsSavingActual(false);
    }
  }

  // loadHistoryData fetches forecast and actual history using profile/date filters.
  async function loadHistoryData(override = null) {
    setIsLoadingHistory(true);

    const filter = override || {
      profile_id: historyProfileID || undefined,
      start_date: historyStartDate || undefined,
      end_date: historyEndDate || undefined,
    };

    try {
      const [historyData, actualsData] = await Promise.all([getForecastHistory(filter), getActualHistory(filter)]);

      setForecastHistory(historyData.forecasts || []);
      setActualHistory(actualsData.actuals || []);
    } catch {
      setForecastHistory([]);
      setActualHistory([]);
    } finally {
      setIsLoadingHistory(false);
    }
  }

  // handleApplyHistoryFilter refreshes history lists with selected profile/date scope.
  async function handleApplyHistoryFilter(event) {
    event.preventDefault();
    setError("");
    setFeedback("");

    if (historyStartDate && historyEndDate && historyStartDate > historyEndDate) {
      setError("Start date tidak boleh lebih besar dari end date.");
      return;
    }

    await loadHistoryData();
    setFeedback("Filter history berhasil diterapkan.");
  }

  async function handleUseDeviceLocation() {
    if (!navigator.geolocation) {
      setError("Browser tidak mendukung geolocation.");
      return;
    }

    setIsLocating(true);
    setError("");
    setFeedback("");

    navigator.geolocation.getCurrentPosition(
      (position) => {
        setProfileForm((current) => ({
          ...current,
          lat: position.coords.latitude.toFixed(6),
          lng: position.coords.longitude.toFixed(6),
        }));
        setFeedback("Latitude dan longitude berhasil diisi dari lokasi device.");
        setIsLocating(false);
      },
      () => {
        setError("Gagal mengambil lokasi device. Pastikan izin lokasi diaktifkan.");
        setIsLocating(false);
      },
      {
        enableHighAccuracy: true,
        timeout: 10000,
        maximumAge: 0,
      },
    );
  }

  // loadDevices fetches user-owned devices for integration management screen.
  async function loadDevices() {
    setIsLoadingDevices(true);
    setError("");

    try {
      const result = await listDevices();
      setDevices(result.devices || []);
    } catch (loadError) {
      setError(loadError.message);
      setDevices([]);
    } finally {
      setIsLoadingDevices(false);
    }
  }

  // handleCreateDevice registers one new device and shows one-time api key.
  async function handleCreateDevice(event) {
    event.preventDefault();
    setIsSavingDevice(true);
    setError("");
    setFeedback("");

    try {
      const payload = {
        name: deviceForm.name,
        external_id: deviceForm.external_id,
        ...(deviceForm.solar_profile_id ? { solar_profile_id: deviceForm.solar_profile_id } : {}),
        ...(editingDeviceID ? { is_active: editingDeviceIsActive } : {}),
      };

      if (editingDeviceID) {
        await updateDevice(editingDeviceID, payload);
        setFeedback("Device berhasil diperbarui.");
      } else {
        const result = await createDevice(payload);
        setLatestDeviceKey(result.api_key || "");
        setFeedback("Device berhasil dibuat. Simpan API key sekarang karena hanya tampil sekali.");
      }
      setDeviceForm({ ...emptyDeviceForm, solar_profile_id: profile?.id || "" });
      setEditingDeviceID("");
      setEditingDeviceIsActive(true);
      await loadDevices();
    } catch (submitError) {
      setError(submitError.message);
    } finally {
      setIsSavingDevice(false);
    }
  }

  // handleRotateDeviceKey rotates one device api key and refreshes list.
  async function handleRotateDeviceKey(deviceID) {
    setRotatingDeviceID(deviceID);
    setError("");
    setFeedback("");

    try {
      const result = await rotateDeviceKey(deviceID);
      setLatestDeviceKey(result.api_key || "");
      setFeedback("API key berhasil di-rotate. Simpan key baru sekarang.");
      await loadDevices();
    } catch (rotateError) {
      setError(rotateError.message);
    } finally {
      setRotatingDeviceID("");
    }
  }

  // copyLatestDeviceKey copies one-time key to clipboard.
  async function copyLatestDeviceKey() {
    if (!latestDeviceKey) {
      return;
    }

    try {
      await navigator.clipboard.writeText(latestDeviceKey);
      setFeedback("API key berhasil dicopy.");
    } catch {
      setError("Gagal copy API key. Copy manual dari field.");
    }
  }

  function handleEditDevice(deviceItem) {
    setEditingDeviceID(deviceItem.id);
    setEditingDeviceIsActive(Boolean(deviceItem.is_active));
    setDeviceForm({
      name: deviceItem.name || "",
      external_id: deviceItem.external_id || "",
      solar_profile_id: deviceItem.solar_profile_id || "",
    });
    setCurrentPage("integration");
  }

  function handleCancelDeviceEdit() {
    setEditingDeviceID("");
    setEditingDeviceIsActive(true);
    setDeviceForm({ ...emptyDeviceForm, solar_profile_id: profile?.id || "" });
  }

  async function handleDeleteDevice(deviceID) {
    if (!window.confirm("Hapus device ini? Telemetry terkait akan ikut terhapus.")) {
      return;
    }

    setDeletingDeviceID(deviceID);
    setError("");
    setFeedback("");

    try {
      await deleteDevice(deviceID);
      if (editingDeviceID === deviceID) {
        handleCancelDeviceEdit();
      }
      await loadDevices();
      setFeedback("Device berhasil dihapus.");
    } catch (deleteError) {
      setError(deleteError.message);
    } finally {
      setDeletingDeviceID("");
    }
  }

  function handleNavigate(page) {
    setCurrentPage(page);
    if (typeof window !== "undefined" && window.innerWidth <= 980) {
      setIsSidebarOpen(false);
    }
  }

  if (!token) {
    return (
      <AuthFlow
        authPage={authPage}
        setAuthPage={setAuthPage}
        error={error}
        setError={setError}
        feedback={feedback}
        setFeedback={setFeedback}
        feedbackFading={feedbackFading}
        isAuthLoading={isAuthLoading}
        registerForm={registerForm}
        setRegisterForm={setRegisterForm}
        handleRegister={handleRegister}
        loginForm={loginForm}
        setLoginForm={setLoginForm}
        handleLogin={handleLogin}
        simCapacity={simCapacity}
        setSimCapacity={setSimCapacity}
        simCloudCover={simCloudCover}
        setSimCloudCover={setSimCloudCover}
        simTimePreset={simTimePreset}
        setSimTimePreset={setSimTimePreset}
      />
    );
  }

  const historyByKey = new Map();
  forecastHistory.forEach((item) => {
    const key = `${item.date}__${item.solar_profile_id || "none"}`;
    historyByKey.set(key, {
      date: item.date,
      solar_profile_id: item.solar_profile_id || "",
      predicted_kwh: Number(item.predicted_kwh),
      weather_factor: Number(item.weather_factor),
      efficiency: Number(item.efficiency),
      actual_kwh: null,
      source: null,
    });
  });
  actualHistory.forEach((item) => {
    const key = `${item.date}__${item.solar_profile_id || "none"}`;
    const existing = historyByKey.get(key);
    if (existing) {
      existing.actual_kwh = Number(item.actual_kwh);
      existing.source = item.source;
      return;
    }
    historyByKey.set(key, {
      date: item.date,
      solar_profile_id: item.solar_profile_id || "",
      predicted_kwh: null,
      weather_factor: null,
      efficiency: null,
      actual_kwh: Number(item.actual_kwh),
      source: item.source,
    });
  });

  const historyComparisonRows = Array.from(historyByKey.values())
    .map((row) => {
      const hasBoth = row.predicted_kwh !== null && row.actual_kwh !== null;
      const delta = hasBoth ? row.actual_kwh - row.predicted_kwh : null;
      const deltaPct = hasBoth && row.predicted_kwh > 0 ? (delta / row.predicted_kwh) * 100 : null;
      const accuracy = hasBoth && row.actual_kwh > 0 ? Math.max(0, 100 - (Math.abs(row.predicted_kwh - row.actual_kwh) / row.actual_kwh) * 100) : null;
      return {
        ...row,
        delta,
        deltaPct,
        accuracy,
      };
    })
    .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());

  const matchedRows = historyComparisonRows.filter((row) => row.predicted_kwh !== null && row.actual_kwh !== null);
  const historyKpi = {
    pairedDays: matchedRows.length,
    avgAccuracy: matchedRows.length > 0 ? matchedRows.reduce((sum, row) => sum + (row.accuracy || 0), 0) / matchedRows.length : null,
    avgAbsDelta: matchedRows.length > 0 ? matchedRows.reduce((sum, row) => sum + Math.abs(row.delta || 0), 0) / matchedRows.length : null,
    bestDay: matchedRows.length > 0 ? matchedRows.reduce((best, row) => (best === null || (row.accuracy || 0) > (best.accuracy || 0) ? row : best), null) : null,
  };

  const defaultHistoryRow = historyComparisonRows.find((row) => row.predicted_kwh !== null) || historyComparisonRows[0] || null;
  const selectedHistoryRow = historyComparisonRows.find((row) => getHistoryRowKey(row) === selectedHistoryRowKey) || defaultHistoryRow;
  const selectedHistoryProfile = selectedHistoryRow ? profiles.find((p) => p.id === selectedHistoryRow.solar_profile_id) || null : null;
  const selectedHistoryRisk = selectedHistoryRow ? getWeatherRisk(selectedHistoryRow.weather_factor) : { label: "--", tone: "neutral" };
  const selectedHistoryHourly =
    selectedHistoryRow && selectedHistoryRow.predicted_kwh !== null
      ? getHourlyDistribution(selectedHistoryRow.date, selectedHistoryProfile?.lat, Number(selectedHistoryRow.weather_factor || 0)).map((slot) => ({
          ...slot,
          value: Number(selectedHistoryRow.predicted_kwh) * slot.share,
        }))
      : [];

  return (
    <div className={`app-with-sidebar ${isSidebarOpen ? "sidebar-open" : ""}`}>
      {isSidebarOpen && <button className='sidebar-backdrop' type='button' aria-label='Tutup sidebar' onClick={() => setIsSidebarOpen(false)} />}
      <aside className='sidebar'>
        <div className='sidebar-header'>
          <h1 className='sidebar-title'>Solar Forecast</h1>
          {/* <p className='sidebar-subtitle'>Forecast</p> */}
          <p className='sidebar-subtitle'>by Sinergi IoT Nusantara</p>
          <button className='sidebar-close-button' type='button' onClick={() => setIsSidebarOpen(false)} aria-label='Tutup menu'>
            <FiX />
          </button>
        </div>

        <nav className='sidebar-nav'>
          <button className={`nav-item ${currentPage === "dashboard" ? "active" : ""}`} onClick={() => handleNavigate("dashboard")}>
            <span className='nav-icon'>
              <FiHome />
            </span>
            <span className='nav-label'>Dashboard</span>
          </button>
          <button className={`nav-item ${currentPage === "profile" ? "active" : ""}`} onClick={() => handleNavigate("profile")}>
            <span className='nav-icon'>
              <FiSettings />
            </span>
            <span className='nav-label'>Solar Profile</span>
          </button>
          <button className={`nav-item ${currentPage === "forecast" ? "active" : ""}`} onClick={() => handleNavigate("forecast")}>
            <span className='nav-icon'>
              <FiSun />
            </span>
            <span className='nav-label'>Forecast Hari Ini</span>
          </button>
          <button className={`nav-item ${currentPage === "history" ? "active" : ""}`} onClick={() => handleNavigate("history")}>
            <span className='nav-icon'>
              <FiBarChart2 />
            </span>
            <span className='nav-label'>Forecast History</span>
          </button>
          <button className={`nav-item ${currentPage === "integration" ? "active" : ""}`} onClick={() => handleNavigate("integration")}>
            <span className='nav-icon'>
              <FiCpu />
            </span>
            <span className='nav-label'>Integrasi Device</span>
          </button>
          <button className={`nav-item ${currentPage === "account" ? "active" : ""}`} onClick={() => handleNavigate("account")}>
            <span className='nav-icon'>
              <FiUser />
            </span>
            <span className='nav-label'>Account Info</span>
          </button>
        </nav>

        <div className='sidebar-footer'>
          <button className='logout-button' onClick={() => handleClientLogout(true)}>
            <span className='nav-icon'>
              <FiLogOut />
            </span>
            <span className='nav-label'>Logout</span>
          </button>
        </div>
      </aside>

      <main className='main-content'>
        <div className='content-header'>
          <button className='sidebar-toggle-button' type='button' onClick={() => setIsSidebarOpen((current) => !current)} aria-label='Buka menu'>
            <FiMenu />
          </button>
          <div>
            <p className='page-eyebrow'>{currentUser?.name || "User"}</p>
            <h1 className='page-title'>
              {currentPage === "dashboard" && "Dashboard"}
              {currentPage === "profile" && "Solar Profile"}
              {currentPage === "forecast" && "Forecast Hari Ini"}
              {currentPage === "history" && "Forecast History"}
              {currentPage === "integration" && "Integrasi Device"}
              {currentPage === "account" && "Account Info"}
            </h1>
          </div>
        </div>

        {(error || feedback) && (
          <div className={`banner ${error ? "banner-error" : "banner-success"}${feedbackFading && !error ? " banner-fading" : ""}`}>
            <span className='banner-state-icon' aria-hidden='true'>
              {error ? <FiAlertCircle /> : <FiCheckCircle />}
            </span>
            <span>{error || feedback}</span>
            <button className='banner-close' type='button' onClick={() => (error ? setError("") : setFeedback(""))} aria-label='Tutup'>
              ×
            </button>
          </div>
        )}

        <div className='content-area'>
          {currentPage === "dashboard" && (
            <DashboardSection heartbeatSummary={heartbeatSummary} summary={summary} dashboardForecastHistory={dashboardForecastHistory} dashboardActualHistory={dashboardActualHistory} todayForecast={todayForecastDashboard} activeProfile={profile} />
          )}

          {currentPage === "profile" && (
            <ProfileSection
              profile={profile}
              profiles={profiles}
              profileForm={profileForm}
              setProfileForm={setProfileForm}
              handleCreateProfile={handleCreateProfile}
              isSavingProfile={isSavingProfile}
              editingProfileID={editingProfileID}
              handleEditProfile={handleEditProfile}
              handleDeleteProfile={handleDeleteProfile}
              deletingProfileID={deletingProfileID}
              handleCancelProfileEdit={handleCancelProfileEdit}
              handleUseDeviceLocation={handleUseDeviceLocation}
              isLocating={isLocating}
              setProfile={setProfile}
              setSelectedForecastProfileID={setSelectedForecastProfileID}
            />
          )}

          {currentPage === "forecast" && (
            <ForecastSection
              profiles={profiles}
              profile={profile}
              selectedForecastProfileID={selectedForecastProfileID}
              setSelectedForecastProfileID={setSelectedForecastProfileID}
              isLoadingForecast={isLoadingForecast}
              isLoading={isLoading}
              handleLoadForecast={handleLoadForecast}
              electricityTariff={electricityTariff}
              setElectricityTariff={setElectricityTariff}
              forecast={forecast}
              actualHistory={actualHistory}
              handleRecordActual={handleRecordActual}
              actualForm={actualForm}
              setActualForm={setActualForm}
              isSavingActual={isSavingActual}
            />
          )}

          {currentPage === "history" && (
            <HistorySection
              profiles={profiles}
              historyProfileID={historyProfileID}
              setHistoryProfileID={setHistoryProfileID}
              historyStartDate={historyStartDate}
              setHistoryStartDate={setHistoryStartDate}
              historyEndDate={historyEndDate}
              setHistoryEndDate={setHistoryEndDate}
              handleApplyHistoryFilter={handleApplyHistoryFilter}
              isLoadingHistory={isLoadingHistory}
              historyKpi={historyKpi}
              historyComparisonRows={historyComparisonRows}
              selectedHistoryRow={selectedHistoryRow}
              setSelectedHistoryRowKey={setSelectedHistoryRowKey}
              forecastHistory={forecastHistory}
              actualHistory={actualHistory}
              selectedHistoryProfile={selectedHistoryProfile}
              selectedHistoryRisk={selectedHistoryRisk}
              selectedHistoryHourly={selectedHistoryHourly}
              electricityTariff={electricityTariff}
            />
          )}

          {currentPage === "integration" && (
            <IntegrationSection
              profiles={profiles}
              deviceForm={deviceForm}
              setDeviceForm={setDeviceForm}
              handleCreateDevice={handleCreateDevice}
              isSavingDevice={isSavingDevice}
              latestDeviceKey={latestDeviceKey}
              copyLatestDeviceKey={copyLatestDeviceKey}
              isLoadingDevices={isLoadingDevices}
              devices={devices}
              editingDeviceID={editingDeviceID}
              editingDeviceIsActive={editingDeviceIsActive}
              setEditingDeviceIsActive={setEditingDeviceIsActive}
              handleEditDevice={handleEditDevice}
              handleDeleteDevice={handleDeleteDevice}
              handleCancelDeviceEdit={handleCancelDeviceEdit}
              deletingDeviceID={deletingDeviceID}
              rotatingDeviceID={rotatingDeviceID}
              handleRotateDeviceKey={handleRotateDeviceKey}
            />
          )}

          {currentPage === "account" && (
            <AccountSection
              currentUser={currentUser}
              notificationPreference={notificationPreference}
              setNotificationPreference={setNotificationPreference}
              handleSaveNotificationPreference={handleSaveNotificationPreference}
              isSavingNotificationPreference={isSavingNotificationPreference}
              isLoadingNotificationPreference={isLoadingNotificationPreference}
            />
          )}
        </div>
      </main>
    </div>
  );
}
