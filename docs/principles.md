# Kaiten Principles

> **Kaiten is a research-oriented container runtime built to understand, improve, and experiment with Linux process isolation and cloud-native infrastructure.**

---

# 1. Kernel First

If an abstraction hides an important kernel concept, it is probably the wrong abstraction.

Examples:

* PID Namespace
* UTS Namespace
* Mount Namespace
* cgroups v2
* seccomp
* capabilities
* OverlayFS
* `pivot_root()`

---

# 2. Design for Replacement

Every subsystem should be replaceable.

Examples include:

* Filesystem implementation
* Networking backend
* Image loader
* Scheduler
* Security model
* Storage driver

---

# 3. Measure Before Optimizing

No optimization should exist without evidence.

Every performance claim should be backed by reproducible benchmarks.

Measure:

* Startup latency
* CPU usage
* Memory usage
* Page faults
* Context switches
* Syscall count
* Image pull time
* Build time

---

# 4. Keep Components Small

Each package should own a single responsibility.
Each file should be around 80 lines (no more than 160 lines)
Each package should contain no more than 8 files (excluding tests)
(Unless absolutely necessary)

---

# 5. If deleting an abstraction makes the code simpler without losing clarity, delete it.

---

# 6. Learn From Production Systems

Kaiten should continuously compare itself against production projects such as:

* runc
* containerd
* Docker
* CRI-O
* Podman
* Kubernetes
* Firecracker
* gVisor

The goal is not to imitate them, but to understand their design decisions.

---

# 7. Production-Quality Engineering

Prototype code is acceptable during exploration.

Merged code should meet production-quality standards.

Every subsystem should strive for:

* clean architecture
* comprehensive tests
* clear documentation
* predictable error handling
* maintainability

---

# 8. Documentation Is Part of the Code

Every significant design decision should have an accompanying Architecture Decision Record (ADR).

Documentation should explain:

* the problem,
* alternative approaches,
* chosen solution,
* trade-offs,
* future improvements.

Code explains *how*.

Documentation explains *why*.

---

# 9. Research Is a First-Class Goal

Kaiten is not only a runtime.

It is also a platform for systems research.

Experimental implementations should be encouraged as long as they are:

* isolated,
* measurable,
* documented,
* reproducible.

---

# 10. Security should be a priority

Isolation is meaningless without security.

Every security decision should follow the principle of least privilege.

Rootless execution, capability reduction, seccomp, and other hardening techniques should be considered foundational rather than optional.

---

# Mission

Build a runtime that deepens understanding of Linux, adheres to professional engineering practices, and provides a platform for exploring better approaches to containerization.

Kaiten exists not merely to reproduce existing technology, but to understand it thoroughly enough to improve upon it.
