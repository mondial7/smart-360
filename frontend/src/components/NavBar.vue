<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import { useRouter } from 'vue-router'
import {
  PhChartBar,
  PhChartPieSlice,
  PhUsers,
  PhBuildings,
  PhArrowsClockwise,
  PhClipboardText,
  PhChatCircleText,
  PhSun,
  PhMoon,
  PhList,
  PhX
} from '@phosphor-icons/vue'

const auth = useAuthStore()
const theme = useThemeStore()
const router = useRouter()

const isMobileMenuOpen = ref(false)

function handleLogout() {
  auth.logout()
  router.push('/login')
  closeMobileMenu()
}

function closeMobileMenu() {
  isMobileMenuOpen.value = false
}

function toggleMobileMenu() {
  isMobileMenuOpen.value = !isMobileMenuOpen.value
}

// Close menu when route changes
router.afterEach(() => {
  closeMobileMenu()
})
</script>

<template>
  <div>
    <!-- Mobile Header -->
    <header class="mobile-header">
      <div class="mobile-header__content">
        <router-link to="/" class="mobile-header__brand">Smart 360</router-link>

        <div class="mobile-header__actions">
          <button @click="theme.toggleTheme()" class="mobile-header__theme-btn" aria-label="Toggle theme">
            <PhSun v-if="theme.isDark" :size="24" weight="regular" />
            <PhMoon v-else :size="24" weight="regular" />
          </button>

          <button @click="toggleMobileMenu" class="mobile-header__hamburger" aria-label="Toggle menu">
            <PhList v-if="!isMobileMenuOpen" :size="24" weight="regular" />
            <PhX v-else :size="24" weight="regular" />
          </button>
        </div>
      </div>
    </header>

    <!-- Mobile Overlay -->
    <Transition name="overlay">
      <div v-if="isMobileMenuOpen" class="mobile-overlay" @click="closeMobileMenu"></div>
    </Transition>

    <!-- Mobile Drawer -->
    <Transition name="drawer">
      <aside v-if="isMobileMenuOpen" class="mobile-drawer">
        <div class="mobile-drawer__header">
          <router-link to="/" class="mobile-drawer__brand">Smart 360</router-link>
        </div>

        <nav class="mobile-drawer__nav">
          <router-link to="/" class="nav-link">
            <PhChartBar class="nav-link__icon" :size="20" weight="regular" />
            <span class="nav-link__text">Dashboard</span>
          </router-link>
          <router-link to="/team" class="nav-link">
            <PhUsers class="nav-link__icon" :size="20" weight="regular" />
            <span class="nav-link__text">Team</span>
          </router-link>
          <router-link v-if="auth.isAdmin" to="/teams" class="nav-link">
            <PhBuildings class="nav-link__icon" :size="20" weight="regular" />
            <span class="nav-link__text">Teams</span>
          </router-link>
          <router-link to="/rounds" class="nav-link">
            <PhArrowsClockwise class="nav-link__icon" :size="20" weight="regular" />
            <span class="nav-link__text">Rounds</span>
          </router-link>
          <router-link v-if="auth.isAdmin" to="/analytics" class="nav-link">
            <PhChartPieSlice class="nav-link__icon" :size="20" weight="regular" />
            <span class="nav-link__text">Analytics</span>
          </router-link>
          <router-link v-if="auth.isAdmin" to="/audit-logs" class="nav-link">
            <PhClipboardText class="nav-link__icon" :size="20" weight="regular" />
            <span class="nav-link__text">Audit Log</span>
          </router-link>
          <router-link to="/my-feedback" class="nav-link">
            <PhChatCircleText class="nav-link__icon" :size="20" weight="regular" />
            <span class="nav-link__text">My Feedback</span>
          </router-link>
        </nav>

        <div class="mobile-drawer__footer">
          <div class="mobile-drawer__user">
            <img v-if="auth.user?.photoUrl" :src="auth.user.photoUrl" alt="Profile" class="mobile-drawer__avatar">
            <div class="mobile-drawer__user-info">
              <p class="mobile-drawer__user-name">{{ auth.user?.name }}</p>
              <button @click="handleLogout" class="mobile-drawer__logout">Logout</button>
            </div>
          </div>
        </div>
      </aside>
    </Transition>

    <!-- Desktop Sidebar -->
    <aside class="sidebar">
      <div class="sidebar__header">
        <router-link to="/" class="sidebar__brand">Smart 360</router-link>
      </div>

      <nav class="sidebar__nav">
        <router-link to="/" class="nav-link">
          <PhChartBar class="nav-link__icon" :size="20" weight="regular" />
          <span class="nav-link__text">Dashboard</span>
        </router-link>
        <router-link to="/team" class="nav-link">
          <PhUsers class="nav-link__icon" :size="20" weight="regular" />
          <span class="nav-link__text">Team</span>
        </router-link>
        <router-link v-if="auth.isAdmin" to="/teams" class="nav-link">
          <PhBuildings class="nav-link__icon" :size="20" weight="regular" />
          <span class="nav-link__text">Teams</span>
        </router-link>
        <router-link to="/rounds" class="nav-link">
          <PhArrowsClockwise class="nav-link__icon" :size="20" weight="regular" />
          <span class="nav-link__text">Rounds</span>
        </router-link>
        <router-link v-if="auth.isAdmin" to="/analytics" class="nav-link">
          <PhChartPieSlice class="nav-link__icon" :size="20" weight="regular" />
          <span class="nav-link__text">Analytics</span>
        </router-link>
        <router-link v-if="auth.isAdmin" to="/audit-logs" class="nav-link">
          <PhClipboardText class="nav-link__icon" :size="20" weight="regular" />
          <span class="nav-link__text">Audit Log</span>
        </router-link>
        <router-link to="/my-feedback" class="nav-link">
          <PhChatCircleText class="nav-link__icon" :size="20" weight="regular" />
          <span class="nav-link__text">My Feedback</span>
        </router-link>
      </nav>

      <div class="sidebar__footer">
        <button @click="theme.toggleTheme()" class="sidebar__theme-btn">
          <PhSun v-if="theme.isDark" :size="16" weight="regular" />
          <PhMoon v-else :size="16" weight="regular" />
          <span>{{ theme.isDark ? 'Light' : 'Dark' }}</span>
        </button>

        <div class="sidebar__user">
          <img v-if="auth.user?.photoUrl" :src="auth.user.photoUrl" alt="Profile" class="sidebar__avatar">
          <div class="sidebar__user-details">
            <span class="sidebar__user-name">{{ auth.user?.name }}</span>
            <button class="sidebar__logout" @click="handleLogout">Logout</button>
          </div>
        </div>
      </div>
    </aside>
  </div>
</template>

<style scoped lang="scss">
// Mobile Header
.mobile-header {
  display: flex;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 50;
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);

  @media (min-width: 1024px) {
    display: none;
  }

  &__content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 1rem;
  }

  &__brand {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--color-primary);
    text-decoration: none;
  }

  &__actions {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  &__theme-btn,
  &__hamburger {
    min-height: 44px;
    min-width: 44px;
    padding: 0.5rem;
    border-radius: 0.5rem;
    background: transparent;
    border: none;
    cursor: pointer;
    color: var(--text-primary);
    transition: background 0.2s;

    &:hover {
      background: var(--bg-secondary);
    }
  }

  &__icon {
    width: 1.5rem;
    height: 1.5rem;
  }
}

// Mobile Overlay
.mobile-overlay {
  position: fixed;
  inset: 0;
  z-index: 40;
  background: rgba(0, 0, 0, 0.5);

  @media (min-width: 1024px) {
    display: none;
  }
}

// Mobile Drawer
.mobile-drawer {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 280px;
  z-index: 50;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);

  @media (min-width: 1024px) {
    display: none;
  }

  &__header {
    padding: 1.5rem;
    border-bottom: 1px solid var(--border-color);
  }

  &__brand {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-primary);
    text-decoration: none;
  }

  &__nav {
    flex: 1;
    padding: 1rem 0;
    overflow-y: auto;
  }

  &__footer {
    padding: 1rem;
    border-top: 1px solid var(--border-color);
  }

  &__user {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  &__avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;
  }

  &__user-info {
    flex: 1;
    min-width: 0;
  }

  &__user-name {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-primary);
    display: block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  &__logout {
    font-size: 0.75rem;
    color: var(--text-secondary);
    background: none;
    border: none;
    cursor: pointer;
    padding: 0;
    transition: color 0.2s;

    &:hover {
      color: var(--color-primary);
    }
  }
}

// Desktop Sidebar
.sidebar {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 250px;
  background: var(--bg-primary);
  border-right: 1px solid var(--border-color);
  flex-direction: column;
  z-index: 100;

  @media (min-width: 1024px) {
    display: flex;
  }

  &__header {
    padding: 1.5rem;
    border-bottom: 1px solid var(--border-color);
  }

  &__brand {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--color-primary);
    text-decoration: none;
  }

  &__nav {
    flex: 1;
    padding: 1rem 0;
    overflow-y: auto;
  }

  &__footer {
    padding: 1rem;
    border-top: 1px solid var(--border-color);
  }

  &__theme-btn {
    width: 100%;
    margin-bottom: 0.75rem;
    padding: 0.5rem 1rem;
    border-radius: 0.5rem;
    background: var(--bg-secondary);
    border: none;
    color: var(--text-primary);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    transition: background 0.2s;
    min-height: 44px;

    &:hover {
      background: var(--bg-tertiary);
    }
  }

  &__theme-icon {
    width: 1rem;
    height: 1rem;
  }

  &__user {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  &__avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;
  }

  &__user-details {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    min-width: 0;
  }

  &__user-name {
    font-weight: 500;
    color: var(--text-primary);
    font-size: 0.875rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  &__logout {
    padding: 0.375rem 0.75rem;
    border: 1px solid var(--border-color);
    border-radius: 6px;
    background: var(--bg-primary);
    color: var(--text-secondary);
    cursor: pointer;
    transition: all 0.2s;
    font-size: 0.875rem;
    width: 100%;

    &:hover {
      background: var(--bg-secondary);
    }
  }
}

// Navigation Links (shared between mobile and desktop)
.nav-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: var(--text-secondary);
  text-decoration: none;
  font-weight: 500;
  padding: 0.75rem 1.5rem;
  transition: all 0.2s;
  min-height: 44px;

  &:hover {
    background: var(--bg-secondary);
    color: var(--color-primary);
  }

  &.router-link-active {
    background: var(--bg-secondary);
    color: var(--color-primary);
    border-right: 3px solid var(--color-primary);
  }

  &__icon {
    flex-shrink: 0;
  }

  &__text {
    flex: 1;
  }
}

// Transitions
.drawer-enter-active,
.drawer-leave-active {
  transition: transform 0.3s ease-out;
}

.drawer-enter-from {
  transform: translateX(-100%);
}

.drawer-leave-to {
  transform: translateX(-100%);
}

.overlay-enter-active,
.overlay-leave-active {
  transition: opacity 0.3s ease-out;
}

.overlay-enter-from,
.overlay-leave-to {
  opacity: 0;
}
</style>
