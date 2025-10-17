// Performance optimization utilities for TheBoys Launcher

// Performance monitoring utilities
export const performance = {
  // Mark performance events
  mark: (name: string) => {
    if (window.performance && window.performance.mark) {
      window.performance.mark(name);
    }
  },

  // Measure time between marks
  measure: (name: string, startMark: string, endMark?: string) => {
    if (window.performance && window.performance.measure) {
      try {
        window.performance.measure(name, startMark, endMark);
        const measure = window.performance.getEntriesByName(name, 'measure')[0];
        return measure ? measure.duration : 0;
      } catch (e) {
        console.warn('Performance measurement failed:', e);
        return 0;
      }
    }
    return 0;
  },

  // Get navigation timing
  getNavigationTiming: () => {
    if (window.performance && window.performance.timing) {
      const timing = window.performance.timing;
      return {
        dns: timing.domainLookupEnd - timing.domainLookupStart,
        tcp: timing.connectEnd - timing.connectStart,
        request: timing.responseStart - timing.requestStart,
        response: timing.responseEnd - timing.responseStart,
        dom: timing.domContentLoadedEventEnd - timing.domContentLoadedEventStart,
        load: timing.loadEventEnd - timing.loadEventStart,
        total: timing.loadEventEnd - timing.navigationStart,
      };
    }
    return null;
  },
};

// Debounce utility for search and other frequent operations
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout;
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}

// Throttle utility for scroll and resize events
export function throttle<T extends (...args: any[]) => any>(
  func: T,
  limit: number
): (...args: Parameters<T>) => void {
  let inThrottle: boolean;
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), limit);
    }
  };
}

// Memoize expensive computations
export function memoize<T extends (...args: any[]) => any>(
  fn: T,
  getKey?: (...args: Parameters<T>) => string
): T {
  const cache = new Map<string, ReturnType<T>>();

  return ((...args: Parameters<T>) => {
    const key = getKey ? getKey(...args) : JSON.stringify(args);

    if (cache.has(key)) {
      return cache.get(key);
    }

    const result = fn(...args);
    cache.set(key, result);

    // Limit cache size to prevent memory leaks
    if (cache.size > 100) {
      const firstKey = cache.keys().next().value;
      if (firstKey) {
        cache.delete(firstKey);
      }
    }

    return result;
  }) as T;
}

// Virtual scrolling utility for large lists
export interface VirtualScrollItem {
  id: string;
  height: number;
  [key: string]: any;
}

export class VirtualScrollManager {
  private containerHeight: number;
  private itemHeight: number;
  private scrollTop: number = 0;
  private items: VirtualScrollItem[] = [];

  constructor(containerHeight: number, itemHeight: number) {
    this.containerHeight = containerHeight;
    this.itemHeight = itemHeight;
  }

  setItems(items: VirtualScrollItem[]) {
    this.items = items;
  }

  setScrollTop(scrollTop: number) {
    this.scrollTop = scrollTop;
  }

  getVisibleItems() {
    const startIndex = Math.floor(this.scrollTop / this.itemHeight);
    const endIndex = Math.min(
      startIndex + Math.ceil(this.containerHeight / this.itemHeight) + 1,
      this.items.length
    );

    return {
      items: this.items.slice(startIndex, endIndex),
      startIndex,
      endIndex,
      offsetY: startIndex * this.itemHeight,
      totalHeight: this.items.length * this.itemHeight,
    };
  }
}

// Intersection Observer for lazy loading
export const createIntersectionObserver = (
  callback: IntersectionObserverCallback,
  options?: IntersectionObserverInit
) => {
  if ('IntersectionObserver' in window) {
    return new IntersectionObserver(callback, {
      rootMargin: '50px',
      ...options,
    });
  }
  return null;
};

// Image lazy loading
export const lazyLoadImage = (img: HTMLImageElement, src: string) => {
  const observer = createIntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        const target = entry.target as HTMLImageElement;
        target.src = src;
        target.classList.remove('lazy-loading');
        observer?.unobserve(target);
      }
    });
  });

  if (observer) {
    img.classList.add('lazy-loading');
    observer.observe(img);
  } else {
    // Fallback for browsers that don't support IntersectionObserver
    img.src = src;
  }
};

// Preload critical resources
export const preloadResource = (url: string, as: string) => {
  const link = document.createElement('link');
  link.rel = 'preload';
  link.href = url;
  link.as = as;
  document.head.appendChild(link);
};

// Prefetch non-critical resources
export const prefetchResource = (url: string) => {
  const link = document.createElement('link');
  link.rel = 'prefetch';
  link.href = url;
  document.head.appendChild(link);
};

// Service Worker registration for caching
export const registerServiceWorker = async () => {
  if ('serviceWorker' in navigator) {
    try {
      const registration = await navigator.serviceWorker.register('/sw.js');
      console.log('Service Worker registered:', registration);
      return registration;
    } catch (error) {
      console.warn('Service Worker registration failed:', error);
      return null;
    }
  }
  return null;
};

// Memory usage monitoring
export const getMemoryUsage = () => {
  if ('memory' in performance) {
    const memory = (performance as any).memory;
    return {
      used: Math.round(memory.usedJSHeapSize / 1048576), // MB
      total: Math.round(memory.totalJSHeapSize / 1048576), // MB
      limit: Math.round(memory.jsHeapSizeLimit / 1048576), // MB
    };
  }
  return null;
};

// FPS monitoring
export class FPSMonitor {
  private frames: number[] = [];
  private lastTime: number = performance.now();
  private animationId: number | null = null;

  start() {
    const measure = () => {
      const now = performance.now();
      const delta = now - this.lastTime;
      const fps = 1000 / delta;

      this.frames.push(fps);
      if (this.frames.length > 60) {
        this.frames.shift();
      }

      this.lastTime = now;
      this.animationId = requestAnimationFrame(measure);
    };

    this.animationId = requestAnimationFrame(measure);
  }

  stop() {
    if (this.animationId) {
      cancelAnimationFrame(this.animationId);
      this.animationId = null;
    }
  }

  getFPS() {
    if (this.frames.length === 0) return 0;
    const sum = this.frames.reduce((a, b) => a + b, 0);
    return Math.round(sum / this.frames.length);
  }
}

// Bundle size monitoring
export const logBundleSize = () => {
  const resources = performance.getEntriesByType('resource') as PerformanceResourceTiming[];
  const bundleSize = resources
    .filter(resource => resource.name.includes('.js') || resource.name.includes('.css'))
    .reduce((total, resource) => total + (resource.transferSize || 0), 0);

  console.log(`Bundle size: ${Math.round(bundleSize / 1024)} KB`);
};

// Error boundary performance logging
export const logError = (error: Error, errorInfo: any) => {
  console.group('🚨 Performance Error');
  console.error('Error:', error);
  console.error('Error Info:', errorInfo);
  console.log('Memory Usage:', getMemoryUsage());
  console.log('Navigation Timing:', performance.getNavigationTiming());
  console.groupEnd();
};

// Performance metrics collector
export const collectPerformanceMetrics = () => {
  const metrics = {
    memory: getMemoryUsage(),
    navigation: performance.getNavigationTiming(),
    resources: performance.getEntriesByType('resource').length,
    timestamp: new Date().toISOString(),
  };

  // Send metrics to analytics service (optional)
  if (process.env.NODE_ENV === 'production') {
    // Send to your analytics service
    console.log('Performance metrics:', metrics);
  }

  return metrics;
};