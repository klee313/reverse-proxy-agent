package com.rpa.android.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

private val LightColors = lightColorScheme(
    primary = Slate900,
    onPrimary = Color.White,
    secondary = Color(0xFF0EA5E9),
    onSecondary = Color(0xFF0B1120),
    tertiary = Emerald400,
    error = Rose400,
    background = Color(0xFFF8F6F0),
    surface = Color(0xFFFFFEFC),
    surfaceVariant = Color(0xFFEDE9E2),
    onSurface = Slate900,
    onSurfaceVariant = Color(0xFF475569),
)

private val DarkColors = darkColorScheme(
    primary = Color(0xFFE2E8F0),
    onPrimary = Color(0xFF0B1120),
    secondary = Color(0xFF0EA5E9),
    onSecondary = Color(0xFF0B1120),
    tertiary = Emerald400,
    error = Rose400,
    background = Color(0xFF0B1120),
    surface = Color(0xFF0F172A),
    surfaceVariant = Color(0xFF1E293B),
    onSurface = Color(0xFFE2E8F0),
    onSurfaceVariant = Color(0xFF94A3B8),
)

@Composable
fun RpaTheme(content: @Composable () -> Unit) {
    MaterialTheme(
        colorScheme = LightColors,
        typography = RpaTypography,
        content = content
    )
}
