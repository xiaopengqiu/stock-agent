<template>
  <span class="animated-counter" :class="{ 'counting': isCounting }">
    {{ displayValue }}
    <span v-if="suffix" class="suffix">{{ suffix }}</span>
  </span>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'

const props = defineProps({
  value: {
    type: Number,
    required: true
  },
  duration: {
    type: Number,
    default: 1000
  },
  decimalPlaces: {
    type: Number,
    default: 0
  },
  suffix: {
    type: String,
    default: ''
  },
  autoStart: {
    type: Boolean,
    default: true
  }
})

const currentValue = ref(0)
const isCounting = ref(false)
let animationId = null

const displayValue = computed(() => {
  return currentValue.value.toFixed(props.decimalPlaces)
})

function animateCounter(targetValue) {
  if (animationId) {
    cancelAnimationFrame(animationId)
  }

  const startValue = currentValue.value
  const startTime = performance.now()
  const duration = props.duration

  isCounting.value = true

  function animate(currentTime) {
    const elapsed = currentTime - startTime
    const progress = Math.min(elapsed / duration, 1)

    // 使用缓动函数
    const easeOutQuart = 1 - Math.pow(1 - progress, 4)
    currentValue.value = startValue + (targetValue - startValue) * easeOutQuart

    if (progress < 1) {
      animationId = requestAnimationFrame(animate)
    } else {
      currentValue.value = targetValue
      isCounting.value = false
    }
  }

  animationId = requestAnimationFrame(animate)
}

watch(() => props.value, (newVal) => {
  animateCounter(newVal)
})

onMounted(() => {
  if (props.autoStart) {
    animateCounter(props.value)
  }
})

// 暴露方法
defineExpose({
  start: () => animateCounter(props.value),
  reset: () => {
    currentValue.value = 0
    isCounting.value = false
  }
})
</script>

<style scoped>
.animated-counter {
  display: inline-flex;
  align-items: baseline;
  font-variant-numeric: tabular-nums;
  transition: color 0.3s ease;
}

.animated-counter.counting {
  animation: pulse 0.5s ease-in-out;
}

.suffix {
  margin-left: 2px;
  font-size: 0.8em;
  opacity: 0.7;
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}
</style>
