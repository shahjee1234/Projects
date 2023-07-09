package com.example.springbootbackend.service;



import com.example.springbootbackend.dto.CreateUserReq;
import com.example.springbootbackend.model.Employee;
//import com.example.springbootbackend.repository.AppUserRepository;
import com.example.springbootbackend.repository.EmployeeRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.core.GrantedAuthority;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.userdetails.User;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.core.userdetails.UserDetailsService;
import org.springframework.security.core.userdetails.UsernameNotFoundException;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

import java.util.ArrayList;

@Service
public class UserDetailsServiceImpl implements UserDetailsService {
    @Autowired
    private EmployeeRepository employeeRepository;

    @Autowired
    private PasswordEncoder passwordEncoder;

    @Override
    public UserDetails loadUserByUsername(String username) throws UsernameNotFoundException {
        // Find user by username
        Employee employee = employeeRepository.findByUsername(username)
                .orElseThrow(() -> new UsernameNotFoundException(username + " not found."));

        var authorities = new ArrayList<GrantedAuthority>();
        authorities.add(new SimpleGrantedAuthority(employee.getUserRole()));

        return new User(employee.getUsername(), employee.getPassword(), authorities);
    }

    public Employee createUser(CreateUserReq req) throws Exception {
        // Extract parameters from request
        String username = req.getUsername();
        String password = req.getPassword();
        String userRole = req.getUserRole();

        // Check whether username exists or not
        boolean isExists = employeeRepository.existsByUsername(username);

        if (isExists) {
            throw new Exception("User already exists.");
        }

        // Create new user
        Employee employee = new Employee();
        employee.setUsername(username);
        employee.setPassword(passwordEncoder.encode(password));
        employee.setUserRole(userRole);

        // Save user
        return employeeRepository.save(employee);
    }
}

