"use client";

import { motion } from "framer-motion";
import { 
  Sun, 
  ArrowRight, 
  BarChart3, 
  ShieldCheck, 
  Zap, 
  Activity, 
  Server, 
  CheckCircle2 
} from "lucide-react";
import Link from "next/link";
import { cn } from "@/lib/utils";

export default function LandingPage() {
  return (
    <div className="relative min-h-screen overflow-hidden bg-background">
      {/* Background Decorative Elements */}
      <div className="absolute top-0 -left-4 w-72 h-72 bg-primary/10 rounded-full blur-3xl" />
      <div className="absolute bottom-0 -right-4 w-96 h-96 bg-secondary/10 rounded-full blur-3xl" />

      {/* Navigation */}
      <nav className="relative z-10 flex items-center justify-between px-6 py-6 mx-auto max-w-7xl">
        <div className="flex items-center gap-2">
          <div className="flex items-center justify-center w-10 h-10 rounded-xl bg-primary shadow-premium">
            <Sun className="text-white w-6 h-6" />
          </div>
          <span className="text-xl font-bold tracking-tight">SolarForecast</span>
        </div>
        <div className="hidden md:flex items-center gap-8 text-sm font-medium text-muted-foreground">
          <Link href="#features" className="hover:text-primary transition-colors">Features</Link>
          <Link href="#analytics" className="hover:text-primary transition-colors">Analytics</Link>
          <Link href="#pricing" className="hover:text-primary transition-colors">Pricing</Link>
        </div>
        <Link 
          href="/login" 
          className="px-5 py-2.5 text-sm font-semibold text-white bg-primary rounded-xl shadow-premium hover:opacity-90 transition-all"
        >
          Sign In
        </Link>
      </nav>

      {/* Hero Section */}
      <main className="relative z-10 px-6 pt-20 pb-32 mx-auto max-w-7xl">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <motion.div
            initial={{ opacity: 0, x: -30 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.8, ease: "easeOut" }}
            className="text-left"
          >
            <span className="inline-block px-4 py-1.5 mb-6 text-xs font-bold tracking-wider text-primary uppercase bg-primary/10 rounded-full border border-primary/20">
              Next-Gen Energy Intelligence
            </span>
            <h1 className="text-5xl md:text-7xl font-black tracking-tight mb-8 leading-[1.1] text-foreground">
              Predict. Monitor. <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-emerald-500 to-teal-600">Optimize Solar.</span>
            </h1>
            <p className="max-w-xl mb-10 text-lg text-muted-foreground leading-relaxed font-medium">
              High-precision forecasting for solar energy production. 
              Empower your green energy transition with AI-driven analytics and 
              real-time device monitoring.
            </p>
            <div className="flex flex-col sm:flex-row items-center gap-4">
              <Link 
                href="/register" 
                className="w-full sm:w-auto flex items-center justify-center gap-2 px-8 py-4 text-lg font-bold text-white bg-primary rounded-2xl shadow-premium hover:translate-y-[-2px] transition-all group"
              >
                Get Started for Free
                <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
              </Link>
              <Link 
                href="#demo" 
                className="w-full sm:w-auto px-8 py-4 text-lg font-bold text-foreground bg-card border border-border rounded-2xl hover:bg-accent transition-all text-center"
              >
                View Live Demo
              </Link>
            </div>
            
            <div className="mt-12 flex items-center gap-6">
              <div className="flex -space-x-3">
                {[1,2,3,4].map(i => (
                  <div key={i} className="w-10 h-10 rounded-full border-2 border-background bg-accent flex items-center justify-center text-[10px] font-bold">
                    U{i}
                  </div>
                ))}
              </div>
              <p className="text-sm text-muted-foreground font-bold">
                Trusted by <span className="text-foreground">500+</span> solar farm operators
              </p>
            </div>
          </motion.div>

          {/* 3D Animation Asset Container */}
          <motion.div
            initial={{ opacity: 0, scale: 0.8, rotateY: 20 }}
            animate={{ opacity: 1, scale: 1, rotateY: 0 }}
            transition={{ duration: 1, delay: 0.2 }}
            className="relative perspective-1000"
          >
            <motion.div
              animate={{ 
                y: [0, -20, 0],
                rotate: [0, 1, 0]
              }}
              transition={{ 
                duration: 6, 
                repeat: Infinity, 
                ease: "easeInOut" 
              }}
              className="relative z-10 rounded-[3rem] overflow-hidden shadow-2xl border border-white/10"
            >
              <img 
                src="/3d-hero.png" 
                alt="3D Solar Intelligence" 
                className="w-full h-auto object-cover transform scale-105"
              />
              <div className="absolute inset-0 bg-gradient-to-t from-background/40 to-transparent" />
            </motion.div>

            {/* Decorative Floating Cards */}
            <motion.div
              animate={{ y: [0, 15, 0] }}
              transition={{ duration: 4, repeat: Infinity, ease: "easeInOut", delay: 0.5 }}
              className="absolute -top-8 -right-8 z-20 p-4 bg-card/80 backdrop-blur-xl border border-border rounded-2xl shadow-xl hidden md:block"
            >
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-emerald-500/10 flex items-center justify-center text-emerald-500">
                  <Zap className="w-4 h-4" />
                </div>
                <div>
                  <p className="text-[10px] font-bold uppercase text-muted-foreground">Efficiency</p>
                  <p className="text-sm font-black">+24.8%</p>
                </div>
              </div>
            </motion.div>

            <motion.div
              animate={{ y: [0, -15, 0] }}
              transition={{ duration: 5, repeat: Infinity, ease: "easeInOut", delay: 1 }}
              className="absolute -bottom-10 -left-10 z-20 p-6 bg-card/80 backdrop-blur-xl border border-border rounded-[2rem] shadow-xl hidden md:block"
            >
              <div className="space-y-3">
                <div className="flex items-center justify-between gap-8">
                  <p className="text-xs font-bold">Monthly Forecast</p>
                  <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                </div>
                <div className="flex items-end gap-1 h-12">
                  {[40, 70, 45, 90, 65, 80, 55].map((h, i) => (
                    <div key={i} className="flex-1 bg-primary/20 rounded-t-sm" style={{ height: `${h}%` }} />
                  ))}
                </div>
              </div>
            </motion.div>
          </motion.div>
        </div>

        {/* Features Section */}
        <div id="features" className="mt-32 space-y-12">
          <div className="text-center max-w-3xl mx-auto">
            <h2 className="text-4xl font-bold mb-4">Powerful Features for Modern Energy</h2>
            <p className="text-muted-foreground">Everything you need to manage, monitor, and scale your solar energy production with confidence.</p>
          </div>
          <div className="grid md:grid-cols-3 gap-8">
            {[
              {
                icon: Zap,
                title: "AI Forecasting",
                desc: "High-precision daily production forecasts using hybrid weather-satellite models with up to 98% accuracy."
              },
              {
                icon: BarChart3,
                title: "ESG Analytics",
                desc: "Automated carbon offset tracking and green certification readiness reports for global compliance."
              },
              {
                icon: ShieldCheck,
                title: "IoT Security",
                desc: "Military-grade encryption for all device communications with automated key rotation and auditing."
              },
              {
                icon: Sun,
                title: "Multi-site Management",
                desc: "Unified control for geographically distributed solar farms with localized weather intelligence."
              },
              {
                icon: Activity,
                title: "Real-time Monitoring",
                desc: "Sub-second telemetry ingestion and anomaly detection to prevent downtime and optimize yield."
              },
              {
                icon: Server,
                title: "Developer API",
                desc: "Robust REST API and webhooks to integrate solar intelligence into your existing enterprise stack."
              }
            ].map((feature, idx) => (
              <motion.div
                key={idx}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.6, delay: 0.1 * idx }}
                className="p-8 rounded-[2.5rem] bg-card border border-border hover:border-primary/50 transition-all text-left group"
              >
                <div className="w-14 h-14 mb-6 flex items-center justify-center rounded-2xl bg-primary/10 text-primary group-hover:scale-110 transition-transform">
                  <feature.icon className="w-7 h-7" />
                </div>
                <h3 className="text-xl font-bold mb-3">{feature.title}</h3>
                <p className="text-muted-foreground leading-relaxed">
                  {feature.desc}
                </p>
              </motion.div>
            ))}
          </div>
        </div>

        {/* Analytics / Demo Section */}
        <div id="demo" className="mt-40 space-y-12">
          <div className="text-center max-w-3xl mx-auto mb-16">
            <span className="text-xs font-bold uppercase tracking-widest text-primary bg-primary/10 px-4 py-1.5 rounded-full">Analytics Preview</span>
            <h2 className="text-4xl font-bold mt-6 mb-4">Data-Driven Decisions</h2>
            <p className="text-muted-foreground">Get a birds-eye view of your entire energy portfolio with our interactive ESG dashboard.</p>
          </div>
          
          <motion.div 
            initial={{ opacity: 0, y: 40 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            className="relative rounded-[3rem] border border-border bg-accent/5 overflow-hidden shadow-2xl"
          >
            <div className="p-8 border-b border-border bg-background flex items-center justify-between">
              <div className="flex gap-2">
                <div className="w-3 h-3 rounded-full bg-red-500" />
                <div className="w-3 h-3 rounded-full bg-amber-500" />
                <div className="w-3 h-3 rounded-full bg-emerald-500" />
              </div>
              <div className="px-4 py-1.5 bg-accent rounded-lg text-xs font-medium text-muted-foreground">esg.solarforecast.ai/dashboard</div>
              <div className="w-8" />
            </div>
            <div className="p-8 lg:p-12">
              <div className="grid lg:grid-cols-3 gap-8">
                <div className="lg:col-span-2 space-y-8">
                  <div className="h-64 bg-card border border-border rounded-3xl p-6 flex flex-col justify-between">
                    <div className="flex items-center justify-between">
                      <p className="font-bold">Global Energy Production</p>
                      <div className="flex gap-2">
                        <div className="px-3 py-1 bg-primary/10 text-primary text-[10px] font-bold rounded-full">LIVE</div>
                      </div>
                    </div>
                    <div className="flex items-end gap-2 h-32">
                      {[30, 45, 35, 60, 80, 50, 90, 75, 40, 85, 65, 95].map((h, i) => (
                        <div key={i} className="flex-1 bg-primary/30 rounded-t-lg animate-pulse" style={{ height: `${h}%`, animationDelay: `${i * 0.1}s` }} />
                      ))}
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-6">
                    <div className="p-6 bg-emerald-500/5 border border-emerald-500/20 rounded-3xl">
                      <p className="text-xs font-bold text-emerald-600 uppercase mb-2">Carbon Offset</p>
                      <p className="text-3xl font-black text-emerald-700">1,284 Tons</p>
                    </div>
                    <div className="p-6 bg-blue-500/5 border border-blue-200 rounded-3xl">
                      <p className="text-xs font-bold text-blue-600 uppercase mb-2">Yield Growth</p>
                      <p className="text-3xl font-black text-blue-700">+18.4%</p>
                    </div>
                  </div>
                </div>
                <div className="bg-card border border-border rounded-3xl p-8 space-y-6">
                  <p className="font-bold">Active Sites</p>
                  <div className="space-y-4">
                    {[
                      { name: "Jakarta Main Office", status: "Optimal", val: "94%" },
                      { name: "Bandung Hub", status: "Warning", val: "72%" },
                      { name: "Surabaya Plant", status: "Optimal", val: "98%" },
                      { name: "Bali Eco Resort", status: "Offline", val: "0%" }
                    ].map((site, i) => (
                      <div key={i} className="flex items-center justify-between p-3 rounded-xl bg-accent/20">
                        <div>
                          <p className="text-sm font-bold">{site.name}</p>
                          <p className={cn(
                            "text-[10px] font-bold uppercase",
                            site.status === "Optimal" ? "text-emerald-500" : site.status === "Warning" ? "text-amber-500" : "text-red-500"
                          )}>{site.status}</p>
                        </div>
                        <p className="font-black text-sm">{site.val}</p>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </motion.div>
        </div>

        {/* Pricing Section */}
        <div id="pricing" className="mt-40 space-y-16">
          <div className="text-center max-w-3xl mx-auto">
            <h2 className="text-4xl font-bold mb-4">Simple, Transparent Pricing</h2>
            <p className="text-muted-foreground">Scale your energy intelligence as you grow. Choose the plan that fits your needs.</p>
          </div>
          <div className="grid md:grid-cols-3 gap-8">
            {[
              {
                name: "Free",
                price: "IDR 0",
                desc: "Perfect for personal use and small setups",
                features: ["1 Solar Profile (Site)", "1 IoT Device", "7 Days History", "Daily Forecast", "Email Notifications"],
                btn: "Start for Free",
                popular: false
              },
              {
                name: "Pro",
                price: "IDR 99.000",
                desc: "Best for growing energy portfolios",
                features: ["5 Solar Profiles", "10 IoT Devices", "90 Days Data History", "PDF Reports & PBB", "REC Readiness Report", "CSV Export Support", "Email Notifications"],
                btn: "Upgrade to Pro",
                popular: true
              },
              {
                name: "Enterprise",
                price: "IDR 499.000",
                desc: "Advanced intelligence for large operators",
                features: ["Unlimited Solar Profiles", "Unlimited IoT Devices", "Lifetime Data History", "ESG Dashboard Multi-site", "White-label Logo Branding", "Public Share ESG Link", "External API Access", "Priority Support"],
                btn: "Contact Sales",
                popular: false
              }
            ].map((plan, idx) => (
              <motion.div
                key={idx}
                initial={{ opacity: 0, scale: 0.95 }}
                whileInView={{ opacity: 1, scale: 1 }}
                viewport={{ once: true }}
                className={cn(
                  "relative p-8 rounded-[3rem] border transition-all flex flex-col justify-between",
                  plan.popular ? "bg-primary text-white border-primary shadow-2xl scale-105 z-10" : "bg-card border-border hover:border-primary/30"
                )}
              >
                {plan.popular && (
                  <div className="absolute -top-4 left-1/2 -translate-x-1/2 px-4 py-1 bg-emerald-400 text-emerald-950 text-[10px] font-black uppercase rounded-full">Most Popular</div>
                )}
                <div>
                  <h3 className="text-2xl font-bold mb-2">{plan.name}</h3>
                  <div className="flex items-baseline gap-1 mb-4">
                    <span className="text-4xl font-black">{plan.price}</span>
                    <span className={cn("text-sm font-medium", plan.popular ? "text-white/70" : "text-muted-foreground")}>/month</span>
                  </div>
                  <p className={cn("text-sm mb-8", plan.popular ? "text-white/80" : "text-muted-foreground")}>{plan.desc}</p>
                  <div className="space-y-4 mb-8">
                    {plan.features.map((f, i) => (
                      <div key={i} className="flex items-center gap-3">
                        <CheckCircle2 className={cn("w-4 h-4", plan.popular ? "text-emerald-300" : "text-primary")} />
                        <span className="text-sm font-medium">{f}</span>
                      </div>
                    ))}
                  </div>
                </div>
                <Link 
                  href={plan.name === "Enterprise" ? "mailto:sales@solarforecast.ai" : "/register"}
                  className={cn(
                    "w-full py-4 rounded-2xl font-bold text-center transition-all",
                    plan.popular ? "bg-white text-primary hover:bg-white/90" : "bg-primary text-white hover:opacity-90 shadow-premium"
                  )}
                >
                  {plan.btn}
                </Link>
              </motion.div>
            ))}
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-border mt-40 py-12 px-6">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row items-center justify-between gap-8">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center text-white">
              <Sun className="w-5 h-5" />
            </div>
            <span className="font-bold">SolarForecast</span>
          </div>
          <div className="flex gap-8 text-sm text-muted-foreground font-medium">
            <Link href="#" className="hover:text-primary transition-colors">Privacy Policy</Link>
            <Link href="#" className="hover:text-primary transition-colors">Terms of Service</Link>
            <Link href="/docs" className="hover:text-primary transition-colors">Documentation</Link>
          </div>
          <p className="text-sm text-muted-foreground">© 2026 SolarForecast AI. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
}
