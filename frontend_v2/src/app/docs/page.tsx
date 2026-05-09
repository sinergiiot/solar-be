"use client";

import React from "react";
import { motion } from "framer-motion";
import { 
  Sun, 
  Terminal, 
  Code2, 
  Cpu, 
  ArrowLeft, 
  CheckCircle2, 
  Copy,
  Zap,
  Activity,
  Clock,
  ShieldCheck
} from "lucide-react";
import Link from "next/link";
import { cn } from "@/lib/utils";
import { useToast } from "@/components/providers/toast-provider";

export default function DocsPage() {
  const { showToast } = useToast();

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    showToast("Copied to clipboard!", "success");
  };

  const curlExample = `curl -X POST http://localhost:8085/ingest/telemetry \\
  -H "Content-Type: application/json" \\
  -H "X-Device-Key: YOUR_DEVICE_KEY" \\
  -d '{
    "device_id": "inv-jakarta-01",
    "timestamp": "${new Date().toISOString()}",
    "actual_kwh": 12.5
  }'`;

  const pythonExample = `import requests
import datetime

API_URL = "http://localhost:8085/ingest/telemetry"
DEVICE_KEY = "YOUR_DEVICE_KEY"

data = {
    "device_id": "inv-jakarta-01",
    "timestamp": datetime.datetime.now(datetime.UTC).isoformat(),
    "actual_kwh": 8.75
}

headers = {
    "X-Device-Key": DEVICE_KEY,
    "Content-Type": "application/json"
}

response = requests.post(API_URL, json=data, headers=headers)
print(f"Status: {response.status_code}")`;

  return (
    <div className="min-h-screen bg-background pb-20">
      {/* Navigation */}
      <nav className="border-b border-border bg-card/50 backdrop-blur-xl sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between">
          <div className="flex items-center gap-8">
            <Link href="/" className="flex items-center gap-2">
              <div className="w-10 h-10 rounded-xl bg-primary flex items-center justify-center text-white shadow-premium">
                <Sun className="w-6 h-6" />
              </div>
              <span className="text-xl font-bold tracking-tight">SolarForecast</span>
            </Link>
            <div className="h-6 w-px bg-border hidden md:block" />
            <span className="text-sm font-bold text-muted-foreground uppercase tracking-widest hidden md:block">Documentation</span>
          </div>
          <Link href="/dashboard" className="flex items-center gap-2 text-sm font-bold hover:text-primary transition-colors">
            <ArrowLeft className="w-4 h-4" />
            Back to Dashboard
          </Link>
        </div>
      </nav>

      <main className="max-w-4xl mx-auto px-6 pt-16">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="space-y-12"
        >
          {/* Header */}
          <div className="space-y-4">
            <div className="w-16 h-16 rounded-2xl bg-primary/10 flex items-center justify-center text-primary mb-6">
              <Cpu className="w-8 h-8" />
            </div>
            <h1 className="text-4xl md:text-5xl font-black tracking-tight">IoT Integration Guide</h1>
            <p className="text-xl text-muted-foreground leading-relaxed">
              Connect your solar inverters, sensors, and data loggers to our AI platform in minutes.
            </p>
          </div>

          {/* Quick Start Steps */}
          <section className="grid md:grid-cols-3 gap-6">
            {[
              { icon: Zap, title: "1. Register Device", desc: "Create a device in the dashboard to get your unique key." },
              { icon: Terminal, title: "2. Send Data", desc: "Push telemetry via our simple REST API endpoint." },
              { icon: Activity, title: "3. Monitor Live", desc: "Watch your production data sync in real-time." }
            ].map((step, i) => (
              <div key={i} className="p-6 rounded-3xl bg-card border border-border">
                <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary mb-4">
                  <step.icon className="w-5 h-5" />
                </div>
                <h3 className="font-bold mb-2">{step.title}</h3>
                <p className="text-sm text-muted-foreground">{step.desc}</p>
              </div>
            ))}
          </section>

          {/* API Reference */}
          <section className="space-y-8">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-emerald-500/10 flex items-center justify-center text-emerald-500">
                <Code2 className="w-4 h-4" />
              </div>
              <h2 className="text-2xl font-bold">API Reference</h2>
            </div>

            <div className="p-8 rounded-[2.5rem] bg-card border border-border space-y-6">
              <div className="flex items-center gap-4">
                <span className="px-3 py-1 bg-primary text-white text-xs font-black rounded-lg">POST</span>
                <code className="text-lg font-bold">/ingest/telemetry</code>
              </div>
              
              <div className="space-y-4">
                <p className="font-bold text-sm uppercase tracking-wider text-muted-foreground">Headers</p>
                <div className="space-y-2">
                  <div className="flex items-center justify-between p-4 rounded-xl bg-accent/30 border border-border">
                    <code className="text-sm">X-Device-Key</code>
                    <span className="text-xs text-muted-foreground">Your Secret Device Key (Required)</span>
                  </div>
                  <div className="flex items-center justify-between p-4 rounded-xl bg-accent/30 border border-border">
                    <code className="text-sm">Content-Type</code>
                    <span className="text-xs text-muted-foreground">application/json</span>
                  </div>
                </div>
              </div>

              <div className="space-y-4">
                <p className="font-bold text-sm uppercase tracking-wider text-muted-foreground">Payload Definition</p>
                <div className="overflow-x-auto">
                  <table className="w-full text-left">
                    <thead>
                      <tr className="border-b border-border">
                        <th className="py-4 text-xs font-bold uppercase">Field</th>
                        <th className="py-4 text-xs font-bold uppercase">Type</th>
                        <th className="py-4 text-xs font-bold uppercase">Description</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border">
                      <tr>
                        <td className="py-4"><code className="text-sm text-primary">device_id</code></td>
                        <td className="py-4 text-sm text-muted-foreground">String</td>
                        <td className="py-4 text-sm">Optional identifier for your hardware log.</td>
                      </tr>
                      <tr>
                        <td className="py-4"><code className="text-sm text-primary">timestamp</code></td>
                        <td className="py-4 text-sm text-muted-foreground">ISO8601</td>
                        <td className="py-4 text-sm">Event time (UTC). Defaults to server time.</td>
                      </tr>
                      <tr>
                        <td className="py-4"><code className="text-sm text-primary">actual_kwh</code></td>
                        <td className="py-4 text-sm text-muted-foreground">Float</td>
                        <td className="py-4 text-sm font-bold">The energy produced (Required).</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </section>

          {/* Code Examples */}
          <section className="space-y-6">
            <div className="flex items-center justify-between">
              <h2 className="text-2xl font-bold">Code Examples</h2>
            </div>

            <div className="space-y-8">
              {/* CURL */}
              <div className="space-y-3">
                <div className="flex items-center justify-between px-2">
                  <p className="text-sm font-bold text-muted-foreground flex items-center gap-2">
                    <Terminal className="w-4 h-4" /> Bash / cURL
                  </p>
                  <button 
                    onClick={() => copyToClipboard(curlExample)}
                    className="p-2 hover:bg-accent rounded-lg transition-colors"
                  >
                    <Copy className="w-4 h-4" />
                  </button>
                </div>
                <div className="p-6 rounded-3xl bg-slate-950 text-slate-300 font-mono text-sm overflow-x-auto leading-relaxed shadow-2xl">
                  <pre>{curlExample}</pre>
                </div>
              </div>

              {/* Python */}
              <div className="space-y-3">
                <div className="flex items-center justify-between px-2">
                  <p className="text-sm font-bold text-muted-foreground flex items-center gap-2">
                    <Code2 className="w-4 h-4" /> Python (Requests)
                  </p>
                  <button 
                    onClick={() => copyToClipboard(pythonExample)}
                    className="p-2 hover:bg-accent rounded-lg transition-colors"
                  >
                    <Copy className="w-4 h-4" />
                  </button>
                </div>
                <div className="p-6 rounded-3xl bg-slate-950 text-slate-300 font-mono text-sm overflow-x-auto leading-relaxed shadow-2xl">
                  <pre>{pythonExample}</pre>
                </div>
              </div>
            </div>
          </section>

          {/* Calculation Logic */}
          <section className="p-8 rounded-[2.5rem] bg-primary/5 border border-primary/20 space-y-6">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center text-primary">
                <Zap className="w-4 h-4" />
              </div>
              <h2 className="text-2xl font-bold">How We Calculate Production</h2>
            </div>
            
            <p className="text-muted-foreground leading-relaxed">
              Our platform is designed to work seamlessly with professional power meters and solar inverters that provide <strong>Lifetime Cumulative Energy (kWh Totalizer)</strong> via Modbus or other protocols.
            </p>

            <div className="grid md:grid-cols-2 gap-8 mt-4">
              <div className="space-y-3">
                <p className="font-bold text-sm uppercase text-primary">The Mechanism</p>
                <p className="text-sm text-muted-foreground">
                  Instead of summing every incoming value, we use <strong>Daily Delta Aggregation</strong>. 
                  We calculate your daily yield by subtracting the previous day's maximum cumulative value from today's maximum.
                </p>
              </div>
              <div className="p-4 bg-background rounded-2xl border border-border font-mono text-xs space-y-2">
                <p className="text-primary font-bold">// Logic Formula</p>
                <p>Yield = Max(Today) - Max(Yesterday)</p>
                <p className="text-muted-foreground"># If new device:</p>
                <p>Yield = Max(Today) - Min(Today)</p>
              </div>
            </div>

            <div className="flex items-start gap-3 p-4 bg-emerald-500/5 border border-emerald-500/20 rounded-2xl">
              <CheckCircle2 className="w-5 h-5 text-emerald-500 shrink-0 mt-0.5" />
              <p className="text-sm text-emerald-900/80">
                <strong>Developer Tip:</strong> You don't need to reset your device's counter every day. Just send the raw cumulative kWh value directly from your Modbus register.
              </p>
            </div>
          </section>

          {/* Data Frequency */}
          <section className="space-y-6">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-500">
                <Clock className="w-4 h-4" />
              </div>
              <h2 className="text-2xl font-bold">Recommended Data Frequency</h2>
            </div>

            <div className="grid md:grid-cols-3 gap-6">
              {[
                { 
                  title: "High Resolution", 
                  time: "1 Minute", 
                  desc: "Ideal for real-time monitoring and immediate anomaly detection.",
                  tag: "Enterprise" 
                },
                { 
                  title: "Standard", 
                  time: "5-15 Minutes", 
                  desc: "The recommended balance between data accuracy and network efficiency.",
                  tag: "Recommended" 
                },
                { 
                  title: "Low Power", 
                  time: "30+ Minutes", 
                  desc: "Suitable for remote sites with limited bandwidth or solar-powered loggers.",
                  tag: "Efficiency" 
                }
              ].map((freq, i) => (
                <div key={i} className="p-6 rounded-3xl bg-card border border-border relative overflow-hidden">
                  <div className="absolute top-0 right-0 px-3 py-1 bg-accent text-[10px] font-bold rounded-bl-xl uppercase text-muted-foreground">
                    {freq.tag}
                  </div>
                  <p className="text-2xl font-black mb-1">{freq.time}</p>
                  <p className="text-sm font-bold text-primary mb-3">{freq.title}</p>
                  <p className="text-xs text-muted-foreground leading-relaxed">{freq.desc}</p>
                </div>
              ))}
            </div>
            
            <div className="p-6 rounded-3xl bg-blue-500/5 border border-blue-500/10">
              <p className="text-sm text-blue-900/80 leading-relaxed">
                <strong>Why frequency matters:</strong> While our forecasting engine works with daily data, having more frequent data points (e.g., every 5 minutes) allows our AI to better understand your site's specific performance profile and detect shading or soiling issues much faster.
              </p>
            </div>
          </section>

          {/* Troubleshooting */}
          <section className="p-8 rounded-[2.5rem] bg-amber-500/5 border border-amber-500/20">
            <div className="flex items-center gap-3 mb-4">
              <ShieldCheck className="w-6 h-6 text-amber-500" />
              <h2 className="text-xl font-bold">Best Practices</h2>
            </div>
            <ul className="space-y-3">
              {[
                "Use HTTPS in production to keep your Device Key secure.",
                "Send data in 15-minute intervals for optimal forecasting accuracy.",
                "Implement local buffering if the device loses internet connectivity.",
                "Rotate your Device Key immediately if you suspect it has been compromised."
              ].map((text, i) => (
                <li key={i} className="flex items-start gap-3 text-sm text-amber-900/70">
                  <CheckCircle2 className="w-4 h-4 text-amber-500 mt-0.5 shrink-0" />
                  {text}
                </li>
              ))}
            </ul>
          </section>
        </motion.div>
      </main>
    </div>
  );
}
