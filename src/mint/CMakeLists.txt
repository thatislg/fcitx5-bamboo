set( fcitx_bamboo_mint_sources
    bamboomint.cpp
)

add_fcitx5_addon(bamboomint ${fcitx_bamboo_mint_sources})
target_link_libraries(bamboomint Fcitx5::Core Fcitx5::Config Bamboo::Core Pthread::Pthread)
target_include_directories(bamboomint PRIVATE ${PROJECT_BINARY_DIR})
install(TARGETS bamboomint DESTINATION "${CMAKE_INSTALL_LIBDIR}/fcitx5")
fcitx5_translate_desktop_file(mint.conf.in mint.conf)
install(FILES "${CMAKE_CURRENT_BINARY_DIR}/mint.conf" DESTINATION "${CMAKE_INSTALL_DATADIR}/fcitx5/inputmethod")
configure_file(mint-addon.conf.in.in mint-addon.conf.in)
fcitx5_translate_desktop_file("${CMAKE_CURRENT_BINARY_DIR}/mint-addon.conf.in" mint-addon.conf)
install(FILES "${CMAKE_CURRENT_BINARY_DIR}/mint-addon.conf" RENAME bamboomint.conf DESTINATION "${FCITX_INSTALL_PKGDATADIR}/addon")
