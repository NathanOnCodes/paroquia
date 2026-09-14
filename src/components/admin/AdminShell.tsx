"use client";

import { useState } from "react";
import Link from "next/link";
import Box from "@mui/material/Box";
import Drawer from "@mui/material/Drawer";
import AppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import List from "@mui/material/List";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import IconButton from "@mui/material/IconButton";
import Typography from "@mui/material/Typography";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import Avatar from "@mui/material/Avatar";
import Divider from "@mui/material/Divider";
import MenuIcon from "@mui/icons-material/Menu";
import DashboardIcon from "@mui/icons-material/Dashboard";
import PeopleIcon from "@mui/icons-material/People";
import GroupsIcon from "@mui/icons-material/Groups";
import CalendarMonthIcon from "@mui/icons-material/CalendarMonth";
import PaymentsIcon from "@mui/icons-material/Payments";
import AccountBalanceIcon from "@mui/icons-material/AccountBalance";
import MailIcon from "@mui/icons-material/Mail";
import AdminPanelSettingsIcon from "@mui/icons-material/AdminPanelSettings";
import SettingsIcon from "@mui/icons-material/Settings";
import LogoutIcon from "@mui/icons-material/Logout";
import { useRouter } from "next/navigation";
import { ROLE_LABELS } from "@/lib/auth/roles";
import type { Role } from "@/lib/db-types";

const drawerWidth = 260;

export default function AdminShell({
  user,
  children,
}: {
  user: { name: string; email: string; role: Role; siteName: string };
  children: React.ReactNode;
}) {
  const router = useRouter();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const isAdmin = user.role === "admin";

  const items = [
    { href: "/admin/dashboard", label: "Painel", icon: <DashboardIcon />, show: true },
    { href: "/admin/dizimistas", label: "Dizimistas", icon: <PeopleIcon />, show: true },
    { href: "/admin/comunidades", label: "Comunidades", icon: <GroupsIcon />, show: true },
    { href: "/admin/eventos", label: "Eventos", icon: <CalendarMonthIcon />, show: true },
    { href: "/admin/contribuicoes", label: "Contribuições", icon: <PaymentsIcon />, show: true },
    { href: "/admin/transparencia", label: "Transparência", icon: <AccountBalanceIcon />, show: true },
    { href: "/admin/contatos", label: "Mensagens", icon: <MailIcon />, show: true },
    { href: "/admin/usuarios", label: "Usuários", icon: <AdminPanelSettingsIcon />, show: isAdmin },
    { href: "/admin/configuracoes", label: "Configurações", icon: <SettingsIcon />, show: isAdmin },
  ];

  const drawerContent = (
    <Box>
      <Toolbar>
        <Typography variant="h6" sx={{ fontWeight: 700, lineHeight: 1.2 }}>
          {user.siteName}
        </Typography>
      </Toolbar>
      <Divider />
      <List>
        {items
          .filter((i) => i.show)
          .map((item) => (
            <ListItemButton key={item.href} component={Link} href={item.href}>
              <ListItemIcon>{item.icon}</ListItemIcon>
              <ListItemText primary={item.label} />
            </ListItemButton>
          ))}
      </List>
    </Box>
  );

  async function handleLogout() {
    setAnchorEl(null);
    await fetch("/auth/logout", { method: "POST" });
    router.push("/");
    router.refresh();
  }

  return (
    <Box sx={{ display: "flex", minHeight: "100vh" }}>
      <AppBar
        position="fixed"
        sx={{ width: { md: `calc(100% - ${drawerWidth}px)` }, ml: { md: `${drawerWidth}px` } }}
      >
        <Toolbar sx={{ justifyContent: "space-between" }}>
          <IconButton
            color="inherit"
            edge="start"
            onClick={() => setMobileOpen(true)}
            sx={{ display: { md: "none" } }}
          >
            <MenuIcon />
          </IconButton>
          <Typography variant="h6" sx={{ display: { xs: "none", md: "block" } }}>
            Painel administrativo
          </Typography>
          <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
            <Typography variant="body2" sx={{ display: { xs: "none", sm: "block" } }}>
              {user.name}
            </Typography>
            <IconButton onClick={(e) => setAnchorEl(e.currentTarget)}>
              <Avatar sx={{ width: 32, height: 32, fontSize: 14 }}>
                {user.name.charAt(0).toUpperCase()}
              </Avatar>
            </IconButton>
          </Box>
        </Toolbar>
      </AppBar>

      <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={() => setAnchorEl(null)}>
        <MenuItem disabled>
          {user.email} · {ROLE_LABELS[user.role]}
        </MenuItem>
        <Divider />
        <MenuItem onClick={handleLogout}>
          <ListItemIcon>
            <LogoutIcon fontSize="small" />
          </ListItemIcon>
          Sair
        </MenuItem>
      </Menu>

      <Box component="nav" sx={{ width: { md: drawerWidth }, flexShrink: { md: 0 } }}>
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={() => setMobileOpen(false)}
          ModalProps={{ keepMounted: true }}
          sx={{ display: { xs: "block", md: "none" }, "& .MuiDrawer-paper": { width: drawerWidth } }}
        >
          {drawerContent}
        </Drawer>
        <Drawer
          variant="permanent"
          open
          sx={{ display: { xs: "none", md: "block" }, "& .MuiDrawer-paper": { width: drawerWidth } }}
        >
          {drawerContent}
        </Drawer>
      </Box>

      <Box component="main" sx={{ flexGrow: 1, p: 3, pt: 10, bgcolor: "background.default", minHeight: "100vh" }}>
        {children}
      </Box>
    </Box>
  );
}
