cmake_minimum_required(VERSION 3.10)
project(matching_engine)
set(CMAKE_CXX_STANDARD 17)
add_executable(matching_engine matching_engine.cpp)
find_package(nlohmann_json 3.2.0 REQUIRED)
find_package(redis++ REQUIRED)
find_package(Crow REQUIRED)
target_link_libraries(matching_engine PRIVATE nlohmann_json::nlohmann_json redis++::redis++ Crow::Crow)
